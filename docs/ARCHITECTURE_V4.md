# Arquitectura v4 - Ponti Backend

## Principios
- FASE 1 = PARIDAD: v4 replica exactamente v3 (mismas columnas/nombres)
- FASE 2 = CLEANUP: Renaming inglés + corrección de bugs
- DRY: Cada cálculo existe en UN solo lugar
- Unicidad: 1 fila por key en cada vista calc

## Capas

```
TABLAS BASE (lots, fields, workorders, supplies, labors)
    │
    ▼
┌─────────────────────────────────────────────────────────┐
│  v4_core (schema)                                       │
│  Funciones math puras SIN acceso a tablas               │
│  - safe_div(numeric, numeric)                           │
│  - per_ha(numeric, numeric)                             │
│  - percentage(numeric, numeric)                         │
└─────────────────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────────────────┐
│  v4_ssot (schema)                                       │
│  FASE 1: Wrappers thin → v3_lot_ssot.*                  │
│  FASE 2: Reimplementación completa                      │
└─────────────────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────────────────┐
│  v4_calc (schema)                                       │
│  Vistas 1-fila-por-entidad (backbone)                   │
│  - lot_base_costs (1 fila por lot_id)                   │
│  - lot_base_income (cuando haya producción)             │
└─────────────────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────────────────┐
│  v4_report (schema)                                     │
│  Contrato estable para Go (paridad con v3)              │
│  - lot_metrics (= v3_lot_metrics)                       │
│  - labor_list (= v3_labor_list)                         │
└─────────────────────────────────────────────────────────┘
```

## Schemas

| Schema | Responsabilidad |
|--------|-----------------|
| v4_core | Funciones math puras (safe_div, per_ha, percentage) |
| v4_ssot | FASE 1: Wrappers → v3_lot_ssot. FASE 2: Reimplementación |
| v4_calc | Vistas 1-fila-por-entidad (lot_base_costs, lot_base_income) |
| v4_report | Contrato estable para Go (paridad con v3) |

## Reglas
1. `COALESCE(value, 0)::numeric` siempre
2. `v4_core.safe_div(a, b)` para divisiones
3. CTEs para evitar alias-cycle
4. 1 fila por key garantizada (blindar con GROUP BY)
5. FASE 1: NO arreglar bugs de naming

## Bugs conocidos (fix en FASE 2)
- `v3_labor_list.usd_cost_ha`: Nombre dice USD pero es ARS
- `v3_labor_list.usd_net_total`: Mismo problema
- Go `labor/repository.go:296`: Doble multiplicación por dólar

## Semántica de rent (000202)
- `rent_per_ha_usd` = arriendo FIJO (expuesta, para UI y active_total)
- `rent_total_per_ha_usd` = arriendo FIJO + % ingresos (NO expuesta, solo para operating_result)

## Dependencias (sin ciclos)

```
tablas base (workorders, supplies, lots)
    │
    ▼
v3_lot_ssot.* (funciones sobre tablas) ← NO depende de vistas
    │
    ▼
v3_workorder_metrics (usa v3_lot_ssot.*) ← NO depende de v3_lot_metrics
    │
    ▼
v4_calc.lot_base_costs (usa v3_workorder_metrics + v4_ssot.*)
    │
    ▼
v4_report.lot_metrics
```
