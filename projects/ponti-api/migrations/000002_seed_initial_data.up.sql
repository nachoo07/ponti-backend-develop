INSERT INTO crops (id, name) VALUES
(1, 'Soja'),
(2, 'Maíz'),
(3, 'Maíz PPM'),
(4, 'Maíz color'),
(5, 'Poroto'),
(6, 'Poroto negro'),
(7, 'Poroto blanco'),
(8, 'Poroto rojo'),
(9, 'Poroto Mung'),
(10, 'Poroto crawberry'),
(11, 'Trigo'),
(12, 'Girasol'),
(13, 'Sorgo'),
(14, 'Garbanzo'),
(15, 'Sésamo');


INSERT INTO lease_types (name) VALUES
('% INGRESO NETO'),
('% UTILIDAD'),
('ARRIENDO FIJO'),
('ARRIENDO FIJO + % INGRESO NETO');

INSERT INTO campaigns (name) VALUES
('2024-2025'),
('2025-2026'),
('2026-2027');

INSERT INTO labor_types (name) VALUES
('Semilla'),
('Agroquímicos'),
('Fertilizantes'),
('Labores');

INSERT INTO labor_categories (name, type_id) VALUES
('Semilla', 1),
('Coadyuvantes', 2),
('Curasemillas', 2),
('Herbicidas', 2),
('Insecticidas', 2),
('Fungicidas', 2),
('Otros Insumos', 2),
('Fertilizantes', 3),
('Siembra', 4),
('Pulverización', 4),
('Otras Labores', 4),
('Riego', 4),
('Cosecha', 4);
