package ar.com.tulocalweb.app.nucleo

import ar.com.tulocalweb.app.BuildConfig
import ar.com.tulocalweb.app.pantallas.tincho.PropuestaTincho
import ar.com.tulocalweb.app.pantallas.tincho.RespuestaTincho
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.flow.flowOn
import kotlinx.coroutines.withContext
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonArray
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonNull
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.booleanOrNull
import kotlinx.serialization.json.doubleOrNull
import kotlinx.serialization.json.jsonPrimitive
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import java.nio.ByteBuffer
import java.nio.CharBuffer
import java.nio.charset.CodingErrorAction
import java.util.concurrent.TimeUnit

/** Un JSON cualquiera: para editar un producto sin perder los campos que la app no muestra (opciones, variantes,
 *  canales) al mandarlo de vuelta (ValorJSON en iOS). */
typealias ValorJSON = JsonElement

val ValorJSON.texto: String? get() = (this as? JsonPrimitive)?.takeIf { it.isString }?.content
val ValorJSON.entero: Int? get() = (this as? JsonPrimitive)?.takeIf { !it.isString }?.doubleOrNull?.toInt()
val ValorJSON.numero: Double? get() = (this as? JsonPrimitive)?.takeIf { !it.isString }?.doubleOrNull
val ValorJSON.bool: Boolean? get() = (this as? JsonPrimitive)?.takeIf { !it.isString }?.booleanOrNull
val ValorJSON.lista: List<ValorJSON>? get() = this as? JsonArray
val ValorJSON.objeto: Map<String, ValorJSON>? get() = this as? JsonObject
operator fun ValorJSON.get(k: String): ValorJSON? = (this as? JsonObject)?.get(k)

/** Para armar un ValorJSON a mano (los .texto, .numero… de iOS). */
object Valor {
    fun texto(s: String): ValorJSON = JsonPrimitive(s)
    fun numero(n: Number): ValorJSON = JsonPrimitive(n)
    fun booleano(b: Boolean): ValorJSON = JsonPrimitive(b)
    fun lista(l: List<ValorJSON>): ValorJSON = JsonArray(l)
    fun objeto(o: Map<String, ValorJSON>): ValorJSON = JsonObject(o)
    val nulo: ValorJSON = JsonNull
}

/** Lo que la app usa de la web (no de la API): subir fotos y hablar con Tincho. La web acepta el mismo token de la
 *  API en esas dos rutas (hooks.server.ts, RUTAS_APP). */
object Web {
    val base: String = BuildConfig.WEB.ifEmpty { "https://tulocalweb.com.ar" }.trimEnd('/')
    private val cliente by lazy { Api.cliente() }
    /** Las pruebas le ponen la API simulada, para que la galería muestre a Tincho respondiendo. */
    internal var clientePrueba: okhttp3.OkHttpClient? = null
    private val http get() = clientePrueba ?: cliente

    /** Sube una foto (JPEG) y devuelve su URL pública (https://tulocalweb.com.ar/f/…). */
    suspend fun subirFoto(datos: ByteArray, token: String?): String = withContext(Dispatchers.IO) {
        val r = Request.Builder().url("$base/panel/fotos").post(datos.toRequestBody("image/jpeg".toMediaType()))
        token?.let { r.header("Authorization", "Bearer $it") }
        @Serializable data class E(val message: String, val plan: String? = null)
        val texto = try {
            http.newCall(r.build()).execute().use {
                val cuerpo = it.body?.string().orEmpty()
                // Lo que dijo la web («Llegaste a los 15 MB de fotos de tu plan…»); 402 es un tope de plan, con
                // el plan que da más espacio.
                if (!it.isSuccessful) {
                    val m = runCatching { Api.json.decodeFromString<E>(cuerpo) }.getOrNull()
                    throw ErrorSubida(m?.message ?: "Probá de nuevo en un rato.", tope = it.code == 402, plan = m?.plan)
                }
                cuerpo
            }
        } catch (e: java.io.IOException) { throw ErrorApi.sinRed }
        @Serializable data class R(val ruta: String)
        val ruta = runCatching { Api.json.decodeFromString<R>(texto).ruta }.getOrNull() ?: throw ErrorApi.desconocido
        if (ruta.startsWith("http")) ruta else base + ruta
    }

    /** El tope de un archivo para descargar (el mismo que pone la web en /panel/archivos). */
    const val maximoArchivo = 20 * 1024 * 1024

    /** Lo que dijo la web al rechazar un archivo («El archivo pesa más de 20 MB.»). */
    class ErrorSubida(
        val mensaje: String,
        /** Es un tope del plan (el espacio de fotos, o la persona quedó en sólo lectura). */
        val tope: Boolean = false,
        /** Con tope: el plan que lo habilita (null en el más grande o en sólo lectura). */
        val plan: String? = null,
    ) : Exception(mensaje)

    /** Sube un archivo para descargar (el premio «descarga»: PDF, EPUB o ZIP de hasta 20 MB, tal cual) y devuelve su
     *  URL pública (https://tulocalweb.com.ar/f/archivos/…). El tipo lo mira la web por los primeros bytes. Si la web lo
     *  rechaza, el error trae su mensaje. */
    suspend fun subirArchivo(datos: ByteArray, token: String?): String = withContext(Dispatchers.IO) {
        val r = Request.Builder().url("$base/panel/archivos").post(datos.toRequestBody("application/octet-stream".toMediaType()))
        token?.let { r.header("Authorization", "Bearer $it") }
        @Serializable data class E(val message: String)
        val largo = http.newBuilder().writeTimeout(120, TimeUnit.SECONDS).readTimeout(120, TimeUnit.SECONDS).callTimeout(120, TimeUnit.SECONDS).build()
        val texto = try {
            largo.newCall(r.build()).execute().use {
                val cuerpo = it.body?.string().orEmpty()
                if (!it.isSuccessful) throw ErrorSubida(runCatching { Api.json.decodeFromString<E>(cuerpo).message }.getOrNull() ?: "Probá de nuevo en un rato.")
                cuerpo
            }
        } catch (e: java.io.IOException) { throw ErrorApi.sinRed }
        @Serializable data class R(val ruta: String)
        val ruta = runCatching { Api.json.decodeFromString<R>(texto).ruta }.getOrNull() ?: throw ErrorApi.desconocido
        if (ruta.startsWith("http")) ruta else base + ruta
    }

    /** El nombre que ve quien lo baja: sin la extensión, hasta 60 letras (nombreDeArchivo en la web). */
    fun nombreDeArchivo(nombre: String): String =
        nombre.replace(Regex("\\.(pdf|epub|zip)$", RegexOption.IGNORE_CASE), "")
            .replace(Regex("[_-]+"), " ").replace(Regex("\\s+"), " ").trim().take(60)

    @Serializable private data class M(val rol: String, val texto: String)
    @Serializable private data class CuerpoBloques(val mensajes: List<M>, val formato: String = "bloques")
    @Serializable private data class Cuerpo(val mensajes: List<M>)

    /** Tincho con bloques (cifras, gráficos, tablas, botones y sugerencias): la respuesta entera en JSON. */
    suspend fun tinchoConBloques(mensajes: List<Pair<String, String>>, token: String?): RespuestaTincho = withContext(Dispatchers.IO) {
        val c = CuerpoBloques(mensajes.takeLast(12).map { M(it.first, it.second.take(1000)) })
        val r = Request.Builder().url("$base/panel/hornero").post(Api.json.encodeToString(c).toRequestBody("application/json".toMediaType()))
        token?.let { r.header("Authorization", "Bearer $it") }
        // Lo que el motor no sabe lo contesta la IA (hasta 4 vueltas): más margen que una pregunta común.
        val largo = http.newBuilder().readTimeout(75, TimeUnit.SECONDS).callTimeout(75, TimeUnit.SECONDS).build()
        val texto = try {
            largo.newCall(r.build()).execute().use { if (!it.isSuccessful) throw ErrorApi.desconocido; it.body?.string().orEmpty() }
        } catch (e: java.io.IOException) { throw ErrorApi.sinRed }
        runCatching { Api.json.decodeFromString<RespuestaTincho>(texto) }.getOrNull() ?: throw ErrorApi.desconocido
    }

    @Serializable private data class Confirmar(val herramienta: String, val argumentos: Map<String, ValorJSON>)
    @Serializable private data class ResultadoConfirmado(val url: String? = null)
    @Serializable private data class Confirmado(val ok: Boolean? = null, val error: String? = null, val incierto: Boolean? = null, val resultado: ResultadoConfirmado? = null)

    /** Cómo salió «Confirmar»: hecha (con el link si creó uno), no se pudo (el motivo) o no sabemos (se cortó la red o se
     *  está haciendo): repetirla con la misma clave es seguro. */
    sealed interface Confirmacion {
        data class Hecha(val url: String?) : Confirmacion
        data class Fallo(val motivo: String) : Confirmacion
        data object Incierta : Confirmacion
    }

    /** «Confirmar» una propuesta de Tincho, con su id como Idempotency-Key: repetirla no la hace dos veces. */
    suspend fun confirmarTincho(p: PropuestaTincho, token: String?): Confirmacion = withContext(Dispatchers.IO) {
        val c = Confirmar(p.herramienta, p.argumentos)
        val r = Request.Builder().url("$base/panel/hornero/confirmar").post(Api.json.encodeToString(c).toRequestBody("application/json".toMediaType()))
        token?.let { r.header("Authorization", "Bearer $it") }
        p.id?.let { r.header("Idempotency-Key", it) }
        try {
            http.newCall(r.build()).execute().use {
                val j = runCatching { Api.json.decodeFromString<Confirmado>(it.body?.string().orEmpty()) }.getOrNull()
                when {
                    it.isSuccessful && j?.ok == true -> Confirmacion.Hecha(j.resultado?.url)
                    it.code >= 500 || j?.incierto == true -> Confirmacion.Incierta
                    else -> Confirmacion.Fallo(j?.error ?: "No se pudo hacer. Probá de nuevo en un rato.")
                }
            }
        } catch (e: java.io.IOException) {
            Confirmacion.Incierta
        }
    }

    /** Una pregunta a Tincho: va devolviendo el texto a medida que llega. */
    fun tincho(mensajes: List<Pair<String, String>>, token: String?): Flow<String> = flow {
        val c = Cuerpo(mensajes.takeLast(12).map { M(it.first, it.second) })
        val r = Request.Builder().url("$base/panel/hornero").post(Api.json.encodeToString(c).toRequestBody("application/json".toMediaType()))
        token?.let { r.header("Authorization", "Bearer $it") }
        val largo = http.newBuilder().readTimeout(0, TimeUnit.MILLISECONDS).build()
        largo.newCall(r.build()).execute().use { res ->
            if (!res.isSuccessful) throw ErrorApi.desconocido
            val fuente = res.body?.byteStream() ?: throw ErrorApi.desconocido
            // Se manda cuando el pedazo es UTF-8 válido (no se corta una letra a la mitad).
            val decodificador = Charsets.UTF_8.newDecoder().onMalformedInput(CodingErrorAction.REPORT)
            val buf = ByteArray(1024)
            var acumulado = ByteArray(0)
            while (true) {
                val n = fuente.read(buf)
                if (n < 0) break
                acumulado += buf.copyOf(n)
                val s = runCatching { decodificador.decode(ByteBuffer.wrap(acumulado)).toString() }.getOrNull()
                if (s != null) { emit(s); acumulado = ByteArray(0) }
            }
        }
    }.flowOn(Dispatchers.IO)
}
