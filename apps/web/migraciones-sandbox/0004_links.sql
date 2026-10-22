-- Los links de pago se abren sin la cookie del visitante (los paga otra persona): se buscan por
-- código entre todos los sandboxes.
CREATE INDEX cambios_por_clave ON cambios (entidad, clave);
