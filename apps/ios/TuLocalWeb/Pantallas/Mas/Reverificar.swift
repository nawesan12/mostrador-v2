import SwiftUI

/// «Confirmá que sos vos» antes de cambiar adónde llega la plata (alias, CBU, titular). La API
/// responde 403 «reverificar»; esta hoja pide la contraseña, el código de la app autenticadora o uno
/// que llega al mail (las cuentas de Google, Apple o WhatsApp no tienen contraseña) y devuelve el
/// permiso de 5 minutos para repetir el pedido. Igual que lib/panel/Reverificar.svelte.
struct Reverificar: View {
    @Environment(Estado.self) private var estado
    @Environment(\.dismiss) private var cerrar
    var alConfirmar: (String) -> Void

    private struct Metodos: Decodable { var clave: Bool; var totp: Bool; var email: String }
    private struct Permiso: Decodable { var token: String }
    private struct ConClave: Encodable { var clave: String }
    private struct ConCodigo: Encodable { var codigo: String }

    @State private var metodos: Metodos?
    @State private var conClave = true
    @State private var valor = ""
    @State private var error: String?
    @State private var cargando = false
    @State private var mandado = false
    @FocusState private var foco: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: 18) {
            Grabber().frame(maxWidth: .infinity)
            VStack(alignment: .leading, spacing: 6) {
                Text("Confirmá que sos vos").estilo(.tituloHoja).foregroundStyle(C.texto)
                Text(Formato.sinViudas("Vas a cambiar adónde te llega la plata. Te avisamos por mail con cada cambio."))
                    .estilo(.subhead).foregroundStyle(C.texto2)
                    .fixedSize(horizontal: false, vertical: true)
            }
            if let metodos {
                VStack(alignment: .leading, spacing: 8) {
                    if conClave {
                        Campo(label: "Tu contraseña", error: error != nil, enfocado: foco) {
                            SecureField("", text: $valor)
                                .textContentType(.password)
                                .estilo(.body16).foregroundStyle(C.texto)
                                .focused($foco)
                                .submitLabel(.go)
                                .onSubmit { Task { await confirmar() } }
                        }
                    } else {
                        Text(Formato.sinViudas(mandado ? "El código que te mandamos a \(metodos.email)"
                                               : metodos.totp ? "El código de tu app autenticadora" : "El código que te mandamos por mail"))
                            .estilo(.footnote).foregroundStyle(C.texto2)
                        CampoCodigo(codigo: $valor, error: error != nil, foco: $foco) { Task { await confirmar() } }
                    }
                    if let error { MensajeError(texto: error) }
                }
                Button {
                    Task { await confirmar() }
                } label: {
                    if cargando { ProgressView().tint(.white) } else { Text("Confirmar") }
                }
                .buttonStyle(BotonPrincipal())
                .disabled(cargando || valor.isEmpty)
                Button(conClave ? "Mejor mandame un código por mail" : mandado ? "Mandármelo de nuevo" : "Mandame un código por mail") {
                    conClave = false
                    valor = ""
                    Task { await mandarCodigo() }
                }
                .estilo(.footnoteFuerte).foregroundStyle(C.texto)
                .frame(minHeight: 44)
            } else {
                ProgressView().frame(maxWidth: .infinity, minHeight: 120)
            }
            Spacer(minLength: 0)
        }
        .padding(.horizontal, 20)
        .task {
            let m: Metodos = (try? await estado.api.get("/v1/auth/reverificar")) ?? Metodos(clave: false, totp: false, email: "")
            metodos = m
            conClave = m.clave
            if !m.clave && !m.totp { await mandarCodigo() }
            foco = true
        }
    }

    private func mandarCodigo() async {
        error = nil
        do {
            try await estado.api.hacer("POST", "/v1/auth/reverificar/codigo", [String: String]())
            mandado = true
        } catch let e as ErrorApi {
            error = e.detalle
        } catch {
            self.error = ErrorApi.desconocido.detalle
        }
    }

    private func confirmar() async {
        error = nil
        cargando = true
        defer { cargando = false }
        do {
            let p: Permiso = conClave
                ? try await estado.api.enviar("POST", "/v1/auth/reverificar", ConClave(clave: valor))
                : try await estado.api.enviar("POST", "/v1/auth/reverificar", ConCodigo(codigo: valor))
            Haptico.exito()
            cerrar()
            alConfirmar(p.token)
        } catch let e as ErrorApi {
            Haptico.error()
            error = e.detalle
            valor = ""
        } catch {
            Haptico.error()
            self.error = ErrorApi.desconocido.detalle
        }
    }
}
