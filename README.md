# Ponti API

1. Ejecutar: make stg-build

2. Crear 2 crops, ejemplo:

curl --location 'localhost:8080/api/v1/crops/public/' \
--header 'Content-Type: application/json' \
--data '{
"name":"wheat"
}'

3. Probar los endpoints de la coleccion Soalen.

DROP TABLE IF EXISTS project_investors;
DROP TABLE IF EXISTS project_managers;
DROP TABLE IF EXISTS lots;
DROP TABLE IF EXISTS fields;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS managers;
DROP TABLE IF EXISTS investors;
DROP TABLE IF EXISTS crops;
DROP TABLE IF EXISTS customers;

-- Table: crops
CREATE TABLE crops (
id BIGSERIAL PRIMARY KEY,
name VARCHAR(50) NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table: customers
CREATE TABLE customers (
id BIGSERIAL PRIMARY KEY,
name VARCHAR(100) NOT NULL,
type VARCHAR(100) NOT NULL
);

-- Table: investors
CREATE TABLE investors (
id BIGSERIAL PRIMARY KEY,
name VARCHAR(255) NOT NULL
);

-- Table: managers
CREATE TABLE managers (
id BIGSERIAL PRIMARY KEY,
name VARCHAR(100) NOT NULL,
type VARCHAR(50) NOT NULL
);

-- Table: projects
CREATE TABLE projects (
id BIGSERIAL PRIMARY KEY,
name VARCHAR(100) NOT NULL,
customer_id BIGINT NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
CONSTRAINT fk_project_customer FOREIGN KEY (customer_id)
REFERENCES customers(id)
ON UPDATE CASCADE ON DELETE RESTRICT
);

-- Table: project_managers (join table for many-to-many relation)
CREATE TABLE project_managers (
project_id BIGINT NOT NULL,
manager_id BIGINT NOT NULL,
PRIMARY KEY (project_id, manager_id),
FOREIGN KEY (project_id) REFERENCES projects(id) ON UPDATE CASCADE ON DELETE CASCADE,
FOREIGN KEY (manager_id) REFERENCES managers(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- Table: project_investors
CREATE TABLE project_investors (
project_id BIGINT NOT NULL,
investor_id BIGINT NOT NULL,
percentage DOUBLE PRECISION NOT NULL,
PRIMARY KEY (project_id, investor_id),
FOREIGN KEY (project_id) REFERENCES projects(id) ON UPDATE CASCADE ON DELETE CASCADE,
FOREIGN KEY (investor_id) REFERENCES investors(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- Table: fields
CREATE TABLE fields (
id BIGSERIAL PRIMARY KEY,
project_id BIGINT,
name VARCHAR(100) NOT NULL,
lease_type_id BIGINT NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (project_id) REFERENCES projects(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- Table: lots
CREATE TABLE lots (
id BIGSERIAL PRIMARY KEY,
name VARCHAR(100) NOT NULL,
field_id BIGINT NOT NULL,
previous_crop_id BIGINT NOT NULL,
current_crop_id BIGINT NOT NULL,
hectares DOUBLE PRECISION NOT NULL,
season VARCHAR(20) NOT NULL,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
FOREIGN KEY (field_id) REFERENCES fields(id) ON UPDATE CASCADE ON DELETE RESTRICT,
FOREIGN KEY (previous_crop_id) REFERENCES crops(id) ON UPDATE CASCADE ON DELETE RESTRICT,
FOREIGN KEY (current_crop_id) REFERENCES crops(id) ON UPDATE CASCADE ON DELETE RESTRICT
);

INSERT INTO crops (id, name) VALUES
(1, 'Soja'),
(2, 'Maíz'),
(3, 'Trigo'),
(4, 'Girasol'),
(5, 'Sorgo'),
(6, 'Cebada'),
(7, 'Alfalfa'),
(8, 'Maní'),
(9, 'Centeno'),
(10, 'Avena');

INSERT INTO lease_types (name) VALUES
('% INGRESO NETO'),
('% UTILIDAD'),
('ARRIENDO FIJO'),
('ARRIENDO FIJO + % INGRESO NETO');

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE "crops" TO "cloudrun-service-account@soalen-app.iam";
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE "customers" TO "cloudrun-service-account@soalen-app.iam";
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE "projects" TO "cloudrun-service-account@soalen-app.iam";
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE "project_managers" TO "cloudrun-service-account@soalen-app.iam";
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE "managers" TO "cloudrun-service-account@soalen-app.iam";
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE "project_investors" TO "cloudrun-service-account@soalen-app.iam";
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE "investors" TO "cloudrun-service-account@soalen-app.iam";
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE "fields" TO "cloudrun-service-account@soalen-app.iam";

GRANT USAGE, SELECT ON SEQUENCE fields_id_seq TO "cloudrun-service-account@soalen-app.iam";

docker build -t us-central1-docker.pkg.dev/pontisoft/ponti-api-registry/ponti-api:0.0.1 .

docker push us-central1-docker.pkg.dev/pontisoft/ponti-api-registry/ponti-api:0.0.1

## **Mapa completo de relaciones entre las entidades**

---

### 1. **Project**

- **Tabla central** del sistema.
- **Relaciones:**

  - **Customer:** Muchos proyectos → un cliente (`CustomerID`)
  - **Campaign:** Muchos proyectos → una campaña (`CampaignID`)
  - **Managers:** Muchos a muchos con managers (tabla intermedia: `project_managers`)
  - **Investors:** Muchos a muchos con inversores (tabla intermedia: `project_investors`)
  - **Fields:** Un proyecto tiene muchos campos/agrocampos

---

### 2. **Customer**

- **Relaciones:**

  - **Projects:** Un cliente tiene muchos proyectos.

- **Campos clave:**

  - `ID`, `Name`, `Type`, ...

---

### 3. **Campaign**

- **Relaciones:**

  - **Projects:** Una campaña tiene muchos proyectos.

- **Campos clave:**

  - `ID`, `Name`, `Year`, ...

---

### 4. **Manager**

- **Relaciones:**

  - **Projects:** Muchos a muchos con proyectos (a través de `project_managers`)

- **Campos clave:**

  - `ID`, `Name`, ...

---

### 5. **Investor**

- **Relaciones:**

  - **Projects:** Muchos a muchos con proyectos (a través de `project_investors`, que además tiene campos como `percentage`)

- **Campos clave:**

  - `ID`, `Name`, `Percentage`, ...

---

### 6. **Field**

- **Relaciones:**

  - **Project:** Un campo pertenece a un proyecto (`ProjectID`)
  - **Lots:** Un campo tiene muchas parcelas/lotes

- **Campos clave:**

  - `ID`, `ProjectID`, `Name`, ...

---

### 7. **Lot**

- **Relaciones:**

  - **Field:** Un lote pertenece a un campo (`FieldID`)
  - **Crop:** Un lote tiene referencia a cultivos (relaciones por `CurrentCropID`, `PreviousCropID`)

- **Campos clave:**

  - `ID`, `FieldID`, `Name`, `CurrentCropID`, `PreviousCropID`, ...

---

### 8. **Crop**

- **Relaciones:**

  - **Lots:** Un cultivo puede estar asociado como actual o previo a varios lotes

- **Campos clave:**

  - `ID`, `Name`, ...

---

## **Diagrama Textual de Relaciones**

```
Customer (1)──────(∞) Project (∞)──────(1) Campaign
                        │
                        │
           ┌────────────┴────────────┐
           │           │             │
        (∞)Manager   (∞)Investor   (∞)Field
           │           │             │
   [project_managers] [project_investors]
                                     │
                                   (∞)
                                   Lot
                                     │
                            ┌────────┴─────────┐
                            │                  │
                      (1)CurrentCrop     (1)PreviousCrop
                            │                  │
                         Crop               Crop
```

## **Explicación de relaciones**

- **Un proyecto** es el núcleo: vincula clientes, campañas, managers, inversores y campos.
- **Managers e Investors**: relación muchos a muchos con proyectos, usando tablas intermedias (que pueden tener campos extra como porcentaje de inversión).
- **Field** (campo): cada proyecto tiene uno o más campos agrícolas.
- **Lot** (lote): cada campo tiene múltiples lotes (parcelas).
- **Crop** (cultivo): un lote puede tener asociado un cultivo actual y uno anterior.

X-API-KEY abc123secreta
X-USER-ID 123


problema migraciones 

1. **Forzamos la migración** a la versión 31: migrate force 31
2. **Ejecutamos la migración** nuevamente: migrate up

