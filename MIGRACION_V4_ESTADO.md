# Migración V4 - Estado y Objetivo

## Resumen Ejecutivo

Migración de vistas de reportes desde schema `public` (v3_*) hacia schema `v4_report` usando el patrón **Strangler Fig**.

**Estado actual:** 95% completado  
**Fecha:** Enero 2025  
**Última actualización:** Investor migrado a v4 + bug aportes reales corregido

---

## 1. ¿Dónde Estamos?

### Vistas Migradas ✅

| Grupo | Vistas | Estado |
|-------|--------|--------|
| **Field Crop** | `field_crop_metrics`, `field_crop_cultivos`, `field_crop_labores`, `field_crop_insumos`, `field_crop_economicos`, `field_crop_rentabilidad` | ✅ Completado |
| **Lot** | `lot_metrics`, `lot_list` | ✅ Completado |
| **Labor** | `labor_metrics`, `labor_list` | ✅ Completado |
| **Workorder** | `workorder_metrics` | ✅ Completado |
| **Summary** | `summary_results` | ✅ Completado (000322) |

### Vistas Pendientes ❌

| Grupo | Vistas | Estado |
|-------|--------|--------|
| **Dashboard** | 5 vistas | ✅ Completado (000325) |
| **Investor** | 4 vistas | ✅ Completado (000326) |

**🎉 Todas las vistas migradas a v4_report**

### Bugs Corregidos ✅

| Bug | Descripción | Migración |
|-----|-------------|-----------|
| Bug del dólar | Labores multiplicaba precio × dólar dos veces | 000317 |
| Arriendo inconsistente | Mostraba valor fijo pero restaba total | 000320 |
| Total Activo | Usaba arriendo fijo en vez del configurado | 000321 |
| Export labores | Decimales como string en vez de float64 | Código Go |
| Summary duplica cálculos | Recalculaba en vez de agregar desde field_crop | 000322 |
| Renta % incorrecta | Numerador/denominador usaban arriendo diferente | 000322 |
| lot_list columnas faltantes | 000318 eliminó sowed_area_ha, harvested_area_ha, etc | 000323 |
| **Summary arriendo fijo** | **Mostraba 119,770 en vez de 161,773 (total)** | **000324** |
| **Performance timeout** | **field_crop_metrics tardaba ~10 seg (5 vistas anidadas)** | **000324** |
| **Stocks: filtro close_date** | query.Where() no se asignaba, mezclaba períodos | Código Go |
| **Stocks: Rubro mostraba Tipo** | Usaba Type.Name en vez de CategoryName | Código Go |
| **Stocks: Consumed = 0** | Faltaba Preload de Category + bug filtro | Código Go |
| **Investor: aportes % acordado** | Mostraba total×%acordado en vez de aportes reales | **000326** |

---

## 2. Arquitectura Actual

```
┌─────────────────────────────────────────────────────────────────┐
│                         CAPA GO                                  │
│  - Lee de vistas v4_report.* (feature flag REPORT_SCHEMA)       │
│  - NO hace cálculos, solo lee                                   │
└─────────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────────┐
│                    SCHEMA: v4_report                            │
│  Vistas de reporte (combinan datos de calc/ssot)                │
│  - field_crop_metrics                                           │
│  - lot_metrics, lot_list                                        │
│  - labor_metrics, labor_list                                    │
│  - workorder_metrics                                            │
└─────────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────────┐
│                    SCHEMAS: v3_*_ssot                           │
│  Funciones SSOT (Single Source of Truth)                        │
│  - v3_lot_ssot.rent_per_ha_for_lot()                           │
│  - v3_lot_ssot.labor_cost_for_lot()                            │
│  - v3_core_ssot.safe_div()                                     │
│  - etc.                                                         │
└─────────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────────┐
│                    SCHEMA: public                               │
│  Tablas de datos                                                │
│  - lots, fields, projects, crops                                │
│  - workorders, labors, supplies                                 │
│  - leases (configuración de arriendo)                          │
└─────────────────────────────────────────────────────────────────┘
```

---

## 3. Objetivo Final

### Estado Final de la BD

```
📁 Schema: v4_report (VISTAS)
├── field_crop_*      ← Ya migrado
├── lot_*             ← Ya migrado
├── labor_*           ← Ya migrado
├── workorder_*       ← Ya migrado
├── dashboard_*       ← Pendiente
└── investor_*        ← Pendiente

📁 Schema: public
├── Tablas de datos (sin cambios)
└── ❌ v3_* vistas (ELIMINADAS en fase Contract)

📁 Schemas SSOT (sin cambios)
├── v3_lot_ssot.*
├── v3_core_ssot.*
└── v3_report_ssot.*
```

### Código Go Final

```go
// Sin feature flag - hardcodeado a v4_report
func (m Model) TableName() string {
    return "v4_report.lot_metrics"
}
```

---

## 4. Fases del Proceso

### FASE 1: Expand ✅ COMPLETADA
- ✅ Crear vistas v4_report.* con paridad exacta a v3
- ✅ Feature flag para switch entre v3/v4
- ✅ Corregir bugs heredados de v3

### FASE 2: Refinar y Validar (EN PROGRESO)

#### 2.1 Migrar Vistas Restantes
- [x] Migrar vistas Dashboard (5 vistas) ✅ 000325
- [ ] Migrar vistas Investor (5 vistas)

#### 2.2 Auditoría de Cálculos Duplicados
- [ ] Identificar cálculos repetidos en vistas
- [ ] Verificar que toda fórmula use funciones SSOT
- [ ] Eliminar cálculos inline que dupliquen SSOT
- [ ] Crear matriz: cálculo → función SSOT única

**Ejemplo de problema:**
```sql
-- ❌ MAL: Vista calcula directamente
SELECT price * quantity AS total

-- ✅ BIEN: Vista usa SSOT
SELECT v3_core_ssot.calculate_total(price, quantity) AS total
```

#### 2.3 Corrección de Naming
- [ ] Columnas: todo en `snake_case` inglés
- [ ] Funciones SSOT: `<accion>_<que>_for_<entidad>`
- [ ] Vistas: `v4_report.<entidad>_<tipo>`
- [ ] Eliminar mezcla español/inglés

**Convenciones objetivo:**
| Elemento | Formato | Ejemplo |
|----------|---------|---------|
| Columna | snake_case EN | `harvested_area_ha` |
| Función SSOT | verbo_noun_for_entity | `rent_per_ha_for_lot()` |
| Vista | schema.entity_type | `v4_report.lot_metrics` |

#### 2.4 Validación de Integridad
- [ ] Verificar tipos de datos consistentes
- [ ] Validar que no haya NULLs inesperados
- [ ] Confirmar precisión decimal (3 decimales mínimo)
- [ ] Test de regresión con datos reales

### FASE 3: Contract (PENDIENTE)
- [ ] Eliminar vistas v3_* del schema public
- [ ] Quitar feature flag del código Go
- [ ] Eliminar código muerto relacionado a v3
- [ ] Documentar arquitectura final

---

## 5. Principios de Arquitectura

### SSOT (Single Source of Truth)
```
❌ MAL:  Vista calcula rent_per_ha con fórmula propia
✅ BIEN: Vista llama v3_lot_ssot.rent_per_ha_for_lot()
```

### No Duplicar Cálculos
```
❌ MAL:  Go calcula: costHaARS = usd_cost_ha × dollar × dollar
✅ BIEN: Go lee: costHaARS directamente de la vista
```

### Consistencia de Valores
```
❌ MAL:  Arriendo mostrado = 100, Arriendo en cálculo = 115
✅ BIEN: Arriendo = 115 en todos lados (según configuración)
```

---

## 6. Próximos Pasos

1. **Estabilizar** (1-2 semanas)
   - Usar la app con v4
   - Detectar bugs si hay

2. ~~**Migrar Dashboard**~~ ✅ Completado
   - ~~Crear vistas v4_report.dashboard_*~~
   - ~~Switch en Go~~

3. **Migrar Investor**
   - Crear vistas v4_report.investor_*
   - Switch en Go

4. **Fase Contract**
   - Eliminar v3_* cuando v4 esté estable
   - Quitar feature flag

---

## 7. Backlog Futuro (Post-Migración)

### Stocks como Vista SQL
**Prioridad:** Baja  
**Beneficio:** Consistencia arquitectónica con el resto del sistema

**Situación actual:**
```go
// Go hace múltiples queries y merge en memoria
1. SELECT * FROM stocks WHERE ...
2. SELECT supply_id, SUM(total_used) FROM workorder_items ...
3. Merge con map[supply_id]consumed
```

**Propuesta futura:**
```sql
CREATE VIEW v4_report.stock_summary AS
SELECT 
    s.*,
    sup.name as supply_name,
    sup.category_name,
    COALESCE(SUM(wi.total_used), 0) as consumed
FROM stocks s
JOIN supplies sup ON sup.id = s.supply_id
LEFT JOIN workorder_items wi ON ...
GROUP BY s.id, sup.id;
```

**Ventajas:**
- SSOT para cálculo de `consumed`
- Menos código Go
- Consistencia con Dashboard/Reports
- 1 query vs 2 queries + merge

**Consideraciones:**
- Stocks tiene CRUD (la tabla sigue siendo necesaria para escrituras)
- Patrón: `stocks` (tabla CRUD) + `stock_summary` (vista lectura)

---

## 8. Feature Flag Actual

```bash
# En GCP Cloud Run
REPORT_SCHEMA=v4_report

# Comportamiento
REPORT_SCHEMA=""          → usa v3_* (legacy)
REPORT_SCHEMA="v4_report" → usa v4_report.* (actual)
```

---

## 9. Migraciones Aplicadas

| Número | Descripción |
|--------|-------------|
| 000301-000316 | Creación inicial de vistas v4 |
| 000317 | Fix bug dólar en labor_list |
| 000318 | Fix arriendo mostrar TOTAL |
| 000320 | Recrear todas las vistas field_crop |
| 000321 | Fix Total Activo usa arriendo configurado |
| 000322 | summary_results agrega desde field_crop (SSOT) |
| 000323 | Fix lot_list columnas faltantes (sowed_area_ha, etc) |
| 000324 | **Reescribe field_crop_metrics** - 1 vista vs 5 anidadas, 10x más rápido, fix arriendo |
| 000325 | **Dashboard migrado a v4** - 5 vistas dashboard_* en v4_report |
| 000326 | **Investor migrado a v4** - 4 vistas investor_* con fix aportes reales |

---

## 10. Optimización Realizada ✅

### field_crop_metrics reescrita (000324)

**Problema original:** `field_crop_metrics` usaba 5 vistas anidadas con 46 llamadas a funciones SSOT, causando timeouts.

**Solución implementada:** Reescribir como UNA sola vista con ~12 llamadas SSOT directas.

**Resultado:**
- ✅ 10x más rápido (~1 seg vs ~10 seg)
- ✅ Bug arriendo corregido (161,773 vs 119,770)
- ✅ `summary_results` ahora agrega correctamente desde `field_crop_metrics`
- ✅ Arquitectura SSOT correcta

**Técnica:** 
```sql
-- ANTES: 5 vistas anidadas
field_crop_metrics → field_crop_economicos → field_crop_cultivos → funciones SSOT

-- AHORA: 1 vista con CTEs
field_crop_metrics AS (
  WITH lot_base AS (SELECT ... funciones SSOT ...),
       aggregated AS (SELECT ... GROUP BY ...)
  SELECT ...
)
```

---

## 11. Checklist de Calidad Final

### Base de Datos ✓

#### Cálculos (SSOT)
- [ ] Cada cálculo tiene UNA sola función SSOT
- [ ] Ninguna vista repite fórmulas
- [ ] Go solo lee, nunca calcula
- [ ] Inventario completo de funciones SSOT

#### Naming (Consistencia)
- [ ] Columnas en inglés snake_case
- [ ] Sin mezcla de idiomas
- [ ] Nombres descriptivos y únicos
- [ ] Prefijos consistentes (v4_report, v3_*_ssot)

#### Tipos de Datos
- [ ] Monetario: `numeric` (nunca float)
- [ ] Porcentajes: `numeric` con precisión
- [ ] IDs: `bigint`
- [ ] Fechas: `timestamp with time zone`

#### Dependencias
- [ ] Vistas documentadas con sus dependencias
- [ ] Orden correcto de creación/eliminación
- [ ] Sin referencias circulares

#### Performance
- [ ] Índices en columnas de JOIN
- [ ] Sin SELECT * en vistas
- [ ] CTEs optimizados

---

## 12. Scripts de Validación a Crear

```
scripts/validation/
├── 01_audit_duplicated_calculations.sql   # Detecta cálculos repetidos
├── 02_audit_naming_conventions.sql        # Verifica nombres
├── 03_audit_ssot_usage.sql                # Verifica uso de SSOT
├── 04_audit_data_types.sql                # Verifica tipos
├── 05_inventory_ssot_functions.sql        # Lista todas las SSOT
└── 06_dependency_graph.sql                # Grafo de dependencias
```

---

## 13. Estado Final Esperado

```
📊 MÉTRICAS OBJETIVO

Cálculos duplicados:     0
Funciones SSOT:         ~30 (una por cálculo)
Vistas v4_report:       ~21 (11 actuales + 10 pendientes)
Vistas v3_*:             0 (eliminadas)
Mezcla de idiomas:       0
Columnas sin tipo:       0
```

### Arquitectura Final Limpia

```
┌─────────────────────────────────────────────────────────────┐
│                      CAPA GO                                │
│   Solo lectura de v4_report.*                               │
│   Sin cálculos, sin transformaciones                        │
└─────────────────────────────────────────────────────────────┘
                           ↓ READ ONLY
┌─────────────────────────────────────────────────────────────┐
│                   v4_report.*                               │
│   Vistas de presentación                                    │
│   Combinan datos ya calculados                              │
└─────────────────────────────────────────────────────────────┘
                           ↓ CALLS
┌─────────────────────────────────────────────────────────────┐
│                   v4_ssot.* / v3_*_ssot.*                   │
│   Funciones SSOT (Single Source of Truth)                   │
│   ÚNICO lugar donde se hacen cálculos                       │
└─────────────────────────────────────────────────────────────┘
                           ↓ READ
┌─────────────────────────────────────────────────────────────┐
│                   public.*                                  │
│   Tablas de datos puros                                     │
│   Sin lógica, sin cálculos                                  │
└─────────────────────────────────────────────────────────────┘
```
