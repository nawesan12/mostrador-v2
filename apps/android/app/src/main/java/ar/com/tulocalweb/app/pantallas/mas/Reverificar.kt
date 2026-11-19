package ar.com.tulocalweb.app.pantallas.mas

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable

// Clon de Pantallas/Mas/Reverificar.swift.

@Serializable private data class MetodosReverificar(val clave: Boolean = false, val totp: Boolean = false, val email: String = "")
@Serializable private data class PermisoReverificar(val token: String)
@Serializable private data class ConClave(val clave: String)
@Serializable private data class ConCodigo(val codigo: String)

/** «Confirmá que sos vos» antes de cambiar adónde llega la plata (alias, CBU, titular). La API responde 403
 *  «reverificar»; esta hoja pide la contraseña, el código de la app autenticadora o uno que llega al mail (las
 *  cuentas de Google o WhatsApp no tienen contraseña) y devuelve el permiso de 5 minutos para repetir el pedido. */
@Composable
fun Reverificar(alConfirmar: (String) -> Unit) {
    val estado = LocalEstado.current
    val cerrar = LocalCerrarHoja.current
    val alcance = rememberCoroutineScope()
    var metodos by remember { mutableStateOf<MetodosReverificar?>(null) }
    var conClave by remember { mutableStateOf(true) }
    var valor by remember { mutableStateOf("") }
    var error by remember { mutableStateOf<String?>(null) }
    var cargando by remember { mutableStateOf(false) }
    var mandado by remember { mutableStateOf(false) }
    var enfocado by remember { mutableStateOf(false) }
    val foco = remember { FocusRequester() }

    suspend fun mandarCodigo() {
        error = null
        try {
            estado.api.hacer("POST", "/v1/auth/reverificar/codigo", emptyMap<String, String>())
            mandado = true
        } catch (e: ErrorApi) { error = e.detalle } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            error = ErrorApi.desconocido.detalle
        }
    }

    suspend fun confirmar() {
        error = null
        cargando = true
        try {
            val p: PermisoReverificar = if (conClave) estado.api.enviar("POST", "/v1/auth/reverificar", ConClave(valor))
            else estado.api.enviar("POST", "/v1/auth/reverificar", ConCodigo(valor))
            Haptico.exito()
            cerrar()
            alConfirmar(p.token)
        } catch (e: ErrorApi) {
            Haptico.error(); error = e.detalle; valor = ""
        } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            Haptico.error(); error = ErrorApi.desconocido.detalle
        } finally { cargando = false }
    }

    LaunchedEffect(Unit) {
        val m = estado.api.opcional<MetodosReverificar>("/v1/auth/reverificar") ?: MetodosReverificar()
        metodos = m
        conClave = m.clave
        if (!m.clave && !m.totp) mandarCodigo()
        runCatching { foco.requestFocus() }
    }

    Column(Modifier.fillMaxWidth().imePadding().padding(horizontal = 20.dp), verticalArrangement = Arrangement.spacedBy(18.dp)) {
        Grabber()
        Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Texto("Confirmá que sos vos", Estilo.tituloHoja, C.texto)
            Texto(Formato.sinViudas("Vas a cambiar adónde te llega la plata. Te avisamos por mail con cada cambio."), Estilo.subhead, C.texto2)
        }
        val m = metodos
        if (m == null) {
            Box(Modifier.fillMaxWidth().heightIn(min = 120.dp), contentAlignment = Alignment.Center) { CircularProgressIndicator(color = C.texto2) }
        } else {
            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                if (conClave) {
                    Campo("Tu contraseña", error != null, enfocado) {
                        Entrada(valor, { valor = it }, "", Modifier.weight(1f).focusRequester(foco), estilo = Estilo.body16, secreto = true,
                            teclado = KeyboardType.Password, accionTeclado = ImeAction.Go, alEnviar = { alcance.launch { confirmar() } },
                            sinCorrector = true, alEnfocar = { enfocado = it })
                    }
                } else {
                    Texto(Formato.sinViudas(if (mandado) "El código que te mandamos a ${m.email}" else if (m.totp) "El código de tu app autenticadora" else "El código que te mandamos por mail"),
                        Estilo.footnote, C.texto2)
                    CampoCodigo(valor, { valor = it }, error != null, foco) { alcance.launch { confirmar() } }
                }
                error?.let { MensajeError(it) }
            }
            BotonPrincipal("Confirmar", { alcance.launch { confirmar() } }, habilitado = !cargando && valor.isNotEmpty(), cargando = cargando)
            Tocable({ conClave = false; valor = ""; alcance.launch { mandarCodigo() } }, Modifier.heightIn(min = 44.dp)) {
                Box(Modifier.heightIn(min = 44.dp), contentAlignment = Alignment.CenterStart) {
                    Texto(if (conClave) "Mejor mandame un código por mail" else if (mandado) "Mandármelo de nuevo" else "Mandame un código por mail",
                        Estilo.footnoteFuerte, C.texto)
                }
            }
        }
    }
}
