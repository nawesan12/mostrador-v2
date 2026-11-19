package ar.com.tulocalweb.app

import ar.com.tulocalweb.app.tema.Simbolos
import org.junit.Assert.assertTrue
import org.junit.Test
import java.io.File

/** Cada nombre de SF Symbol escrito en Icono(), Dato() o AccionCero() se resuelve con la misma regla que
 *  tema/Base.kt › Icono (sin «.fill», o su prefijo): si no, se dibuja la caja de color vacía. */
class SimbolosTest {
    @Test fun cadaIconoTieneSuSimbolo() {
        val literal = Regex("""\b(?:Icono|Dato|AccionCero)\(\s*"([a-z0-9.]+)"""")
        val faltan = File("src/main/java").walkTopDown().filter { it.extension == "kt" }.flatMap { f ->
            literal.findAll(f.readText()).map { it.groupValues[1] to f.name }
        }.filter { (sf, _) ->
            val nombre = sf.replace(".fill", "")
            Simbolos.mapa[nombre] == null && Simbolos.mapa[nombre.substringBeforeLast('.')] == null
        }.toList()
        assertTrue("Sin símbolo en Simbolos.mapa: $faltan", faltan.isEmpty())
    }
}
