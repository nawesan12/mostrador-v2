import Foundation
import UIKit

/// Un JSON cualquiera: para editar un producto sin perder los campos que la app no muestra (opciones,
/// variantes, canales) al mandarlo de vuelta.
enum ValorJSON: Codable, Equatable {
    case texto(String), numero(Double), booleano(Bool), lista([ValorJSON]), objeto([String: ValorJSON]), nulo

    init(from decoder: Decoder) throws {
        let c = try decoder.singleValueContainer()
        if c.decodeNil() { self = .nulo }
        else if let b = try? c.decode(Bool.self) { self = .booleano(b) }
        else if let n = try? c.decode(Double.self) { self = .numero(n) }
        else if let s = try? c.decode(String.self) { self = .texto(s) }
        else if let l = try? c.decode([ValorJSON].self) { self = .lista(l) }
        else { self = .objeto(try c.decode([String: ValorJSON].self)) }
    }

    func encode(to encoder: Encoder) throws {
        var c = encoder.singleValueContainer()
        switch self {
        case let .texto(s): try c.encode(s)
        case let .numero(n):
            if n == n.rounded(), abs(n) < 1e15 { try c.encode(Int(n)) } else { try c.encode(n) }
        case let .booleano(b): try c.encode(b)
        case let .lista(l): try c.encode(l)
        case let .objeto(o): try c.encode(o)
        case .nulo: try c.encodeNil()
        }
    }

    var texto: String? { if case let .texto(s) = self { return s }; return nil }
    var entero: Int? { if case let .numero(n) = self { return Int(n) }; return nil }
    var bool: Bool? { if case let .booleano(b) = self { return b }; return nil }
    var lista: [ValorJSON]? { if case let .lista(l) = self { return l }; return nil }
    subscript(_ k: String) -> ValorJSON? { if case let .objeto(o) = self { return o[k] }; return nil }
}

/// Lo que la app usa de la web (no de la API): subir fotos y hablar con Tincho. La web acepta el
/// mismo token de la API en esas dos rutas (hooks.server.ts, RUTAS_APP).
enum Web {
    static let base = URL(string: "https://tulocalweb.com.ar")!

    /// Sube una foto (JPEG) y devuelve su URL pública (https://tulocalweb.com.ar/f/…).
    static func subirFoto(_ datos: Data, token: String?) async throws -> String {
        var r = URLRequest(url: base.appending(path: "panel/fotos"))
        r.httpMethod = "POST"
        r.setValue("image/jpeg", forHTTPHeaderField: "Content-Type")
        if let token { r.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization") }
        r.httpBody = datos
        let (d, resp) = try await URLSession.shared.data(for: r)
        guard let h = resp as? HTTPURLResponse else { throw ErrorApi.desconocido }
        guard (200..<300).contains(h.statusCode) else {
            // Lo que dijo la web («Llegaste a los 15 MB de fotos de tu plan…»); 402 es un tope de plan.
            struct M: Decodable { var message: String }
            throw ErrorSubida(mensaje: (try? JSONDecoder().decode(M.self, from: d))?.message ?? "Probá de nuevo en un rato.",
                              tope: h.statusCode == 402)
        }
        struct R: Decodable { var ruta: String }
        let ruta = try JSONDecoder().decode(R.self, from: d).ruta
        return ruta.hasPrefix("http") ? ruta : base.absoluteString + ruta
    }

    /// El tope de un archivo para descargar (el mismo que pone la web en /panel/archivos).
    static let maximoArchivo = 20 * 1024 * 1024

    /// Sube un archivo para descargar (el premio «descarga»: PDF, EPUB o ZIP de hasta 20 MB, tal cual) y
    /// devuelve su URL pública (https://tulocalweb.com.ar/f/archivos/…). El tipo lo mira la web por los
    /// primeros bytes. Si la web lo rechaza, el error trae su mensaje.
    static func subirArchivo(_ datos: Data, token: String?) async throws -> String {
        var r = URLRequest(url: base.appending(path: "panel/archivos"), timeoutInterval: 120)
        r.httpMethod = "POST"
        r.setValue("application/octet-stream", forHTTPHeaderField: "Content-Type")
        if let token { r.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization") }
        r.httpBody = datos
        let (d, resp) = try await URLSession.shared.data(for: r)
        struct M: Decodable { var message: String }
        guard let h = resp as? HTTPURLResponse, (200..<300).contains(h.statusCode) else {
            throw ErrorSubida(mensaje: (try? JSONDecoder().decode(M.self, from: d))?.message ?? "Probá de nuevo en un rato.")
        }
        struct R: Decodable { var ruta: String }
        let ruta = try JSONDecoder().decode(R.self, from: d).ruta
        return ruta.hasPrefix("http") ? ruta : base.absoluteString + ruta
    }

    /// Lo que dijo la web al rechazar un archivo («El archivo pesa más de 20 MB.»).
    struct ErrorSubida: Error {
        var mensaje: String
        /// Es un tope del plan (el espacio de fotos, o la persona quedó en sólo lectura).
        var tope = false
    }

    /// El nombre que ve quien lo baja: sin la extensión, hasta 60 letras (nombreDeArchivo en la web).
    static func nombreDeArchivo(_ nombre: String) -> String {
        let sin = nombre.replacingOccurrences(of: #"\.(pdf|epub|zip)$"#, with: "", options: [.regularExpression, .caseInsensitive])
            .replacingOccurrences(of: #"[_-]+"#, with: " ", options: .regularExpression)
            .replacingOccurrences(of: #"\s+"#, with: " ", options: .regularExpression)
            .trimmingCharacters(in: .whitespaces)
        return String(sin.prefix(60))
    }

    /// Una foto elegida, achicada a 1600 px de lado y en JPEG, lista para `subirFoto`.
    static func jpegParaSubir(_ datos: Data) -> Data? {
        guard let img = UIImage(data: datos) else { return nil }
        let escala = min(1, 1600 / max(img.size.width, img.size.height))
        let tam = CGSize(width: img.size.width * escala, height: img.size.height * escala)
        let chica = UIGraphicsImageRenderer(size: tam).image { _ in img.draw(in: CGRect(origin: .zero, size: tam)) }
        return chica.jpegData(compressionQuality: 0.85)
    }

    /// Una pregunta a Tincho: va devolviendo el texto a medida que llega.
    /// Tincho con bloques (cifras, gráficos, tablas, botones y sugerencias): la respuesta entera en JSON.
    static func tinchoConBloques(_ mensajes: [(rol: String, texto: String)], token: String?) async throws -> RespuestaTincho {
        // Lo que el motor no sabe lo contesta la IA (hasta 4 vueltas): más margen que una pregunta común.
        var r = URLRequest(url: base.appending(path: "panel/hornero"), timeoutInterval: 75)
        r.httpMethod = "POST"
        r.setValue("application/json", forHTTPHeaderField: "Content-Type")
        if let token { r.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization") }
        struct M: Encodable { var rol: String; var texto: String }
        struct C: Encodable { var mensajes: [M]; var formato = "bloques" }
        r.httpBody = try JSONEncoder().encode(C(mensajes: mensajes.suffix(12).map { M(rol: $0.rol, texto: String($0.texto.prefix(1000))) }))
        let (d, resp) = try await URLSession.shared.data(for: r)
        guard let h = resp as? HTTPURLResponse, (200..<300).contains(h.statusCode) else { throw ErrorApi.desconocido }
        return try JSONDecoder().decode(RespuestaTincho.self, from: d)
    }

    /// Cómo salió «Confirmar»: hecha (con el link si creó uno), no se pudo (el motivo) o no sabemos
    /// (se cortó la red o se está haciendo): repetirla con la misma clave es seguro.
    enum Confirmacion: Equatable { case hecha(url: String?), error(String), incierta }

    /// «Confirmar» una propuesta de Tincho, con su id como Idempotency-Key: repetirla no la hace dos veces.
    static func confirmarTincho(_ p: PropuestaTincho, token: String?) async -> Confirmacion {
        var r = URLRequest(url: base.appending(path: "panel/hornero/confirmar"), timeoutInterval: 30)
        r.httpMethod = "POST"
        r.setValue("application/json", forHTTPHeaderField: "Content-Type")
        if let token { r.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization") }
        if let id = p.id { r.setValue(id, forHTTPHeaderField: "Idempotency-Key") }
        struct C: Encodable { var herramienta: String; var argumentos: [String: ValorJSON] }
        struct Res: Decodable { var url: String? }
        struct R: Decodable { var ok: Bool?; var error: String?; var incierto: Bool?; var resultado: Res? }
        do {
            r.httpBody = try JSONEncoder().encode(C(herramienta: p.herramienta, argumentos: p.argumentos))
            let (d, resp) = try await URLSession.shared.data(for: r)
            let j = try? JSONDecoder().decode(R.self, from: d)
            let estado = (resp as? HTTPURLResponse)?.statusCode ?? 0
            if (200..<300).contains(estado), j?.ok == true { return .hecha(url: j?.resultado?.url) }
            if estado >= 500 || j?.incierto == true { return .incierta }
            return .error(j?.error ?? "No se pudo hacer. Probá de nuevo en un rato.")
        } catch {
            return .incierta
        }
    }

    static func tincho(_ mensajes: [(rol: String, texto: String)], token: String?) -> AsyncThrowingStream<String, Error> {
        AsyncThrowingStream { cont in
            let tarea = Task {
                do {
                    var r = URLRequest(url: base.appending(path: "panel/hornero"))
                    r.httpMethod = "POST"
                    r.setValue("application/json", forHTTPHeaderField: "Content-Type")
                    if let token { r.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization") }
                    struct M: Encodable { var rol: String; var texto: String }
                    struct C: Encodable { var mensajes: [M] }
                    r.httpBody = try JSONEncoder().encode(C(mensajes: mensajes.suffix(12).map { M(rol: $0.rol, texto: $0.texto) }))
                    let (bytes, resp) = try await URLSession.shared.bytes(for: r)
                    guard let h = resp as? HTTPURLResponse, (200..<300).contains(h.statusCode) else {
                        throw ErrorApi.desconocido
                    }
                    var acumulado = Data()
                    for try await b in bytes {
                        acumulado.append(b)
                        // Se manda cuando el pedazo es UTF-8 válido (no se corta una letra a la mitad).
                        if let s = String(data: acumulado, encoding: .utf8) {
                            cont.yield(s)
                            acumulado.removeAll()
                        }
                    }
                    cont.finish()
                } catch {
                    cont.finish(throwing: error)
                }
            }
            cont.onTermination = { _ in tarea.cancel() }
        }
    }
}
