# Plan de Migración v4

## Fases
- FASE 1: Paridad exacta (solo cambiar TableName en Go)
- FASE 2: Cleanup (renaming + bug fixes)

## Migraciones FASE 1

| # | Migración | Contenido |
|---|-----------|-----------|
| 000300 | create_v4_schemas | v4_core, v4_ssot, v4_calc, v4_report |
| 000301 | create_v4_core_functions | safe_div, per_ha, percentage |
| 000302 | create_v4_ssot_wrappers | Wrappers thin → v3_lot_ssot |
| 000303 | create_v4_calc_lot_base_costs | 1 fila por lot_id |
| 000304 | create_v4_calc_lot_base_income | Para cuando haya producción |
| 000305 | create_v4_report_lot_metrics | Paridad con v3_lot_metrics |
| 000306 | create_v4_report_labor_list | Paridad con v3_labor_list |

## Orden de migración (15 vistas)

| # | Vista v3 | Vista v4 | Semana |
|---|----------|----------|--------|
| 1 | v3_lot_metrics | v4_report.lot_metrics | 1 |
| 2 | v3_lot_list | v4_report.lot_list | 1 |
| 3 | v3_labor_list | v4_report.labor_list | 2 |
| 4 | v3_workorder_metrics | v4_report.workorder_metrics | 2 |
| 5 | v3_workorder_list | v4_report.workorder_list | 2 |
| 6 | v3_report_field_crop_metrics | v4_report.field_crop_metrics | 3 |
| 7 | v3_report_field_crop_labores | v4_report.field_crop_labores | 3 |
| 8 | v3_report_field_crop_insumos | v4_report.field_crop_insumos | 3 |
| 9 | v3_dashboard_metrics | v4_report.dashboard_metrics | 4 |
| 10 | v3_dashboard_management_balance | v4_report.dashboard_management_balance | 4 |
| 11 | v3_dashboard_crop_incidence | v4_report.dashboard_crop_incidence | 4 |
| 12 | v3_dashboard_contributions_progress | v4_report.dashboard_contributions_progress | 5 |
| 13 | v3_dashboard_operational_indicators | v4_report.dashboard_operational_indicators | 5 |
| 14 | v3_report_summary_results_view | v4_report.summary_results | 5 |
| 15 | v3_investor_contribution_data_view | v4_report.investor_contribution | 6 |

## Comandos

```bash
# Aplicar migraciones
make migrate-up

# Validar contrato (schema igual)
make validate-contract

# Validar paridad (datos iguales)
make validate-parity

# Validar todo
make validate-all

# Sanity check (opcional)
make sanity-check
```

## Checklist por vista

### Pre-migración
- [ ] Leer última migración v3
- [ ] Mapear columnas Go model
- [ ] Crear v4 con mismas columnas

### Validación
- [ ] make validate-contract (0 errores)
- [ ] make validate-parity (0 errores)

### Switch Go
- [ ] Cambiar TableName() o Table()
- [ ] NO cambiar nombres de columnas
- [ ] Tests pass

### Post-migración (FASE 2)
- [ ] Renombrar columnas a inglés
- [ ] Corregir bugs
- [ ] Actualizar modelos Go

## Bug documentado (fix en FASE 2)

### Cadena causal del bug del dólar

| Paso | Archivo | Línea | Valor | Problema |
|------|---------|-------|-------|----------|
| 1 | 000089_fix_labor_calculations_v3.up.sql | 86 | `lb.price AS cost_per_ha` | 50 USD/ha ✓ |
| 2 | 000089_fix_labor_calculations_v3.up.sql | 90 | `dollar_average_month` | 1000 ARS/USD ✓ |
| 3 | 000089_fix_labor_calculations_v3.up.sql | 97 | `usd_cost_ha = price × dollar` | 50000 **ARS** (nombre incorrecto) |
| 4 | labor/repository/models/labor_item.go | 34 | `USDCostHa` | 50000 ARS |
| 5 | labor/repository.go | 296 | `costHaARS = USDCostHa × USDAvgValue` | **50,000,000 ARS** ← BUG |

### Fix propuesto (FASE 2)
1. SQL: Renombrar `usd_cost_ha` → `cost_per_ha_ars`
2. Go: Eliminar línea 296, usar valor directo de la vista
