# Scripts de Validación de Paridad

## Uso

```bash
# Validar contrato (schema igual)
make validate-contract

# Validar paridad de datos
make validate-parity

# Validar todo
make validate-all

# Una vista específica
psql -v ON_ERROR_STOP=1 -d ponti -f scripts/validation/parity_v3_lot_metrics.sql
```

## Tipos de checks

### Contract (schema)
- MISSING_IN_V4: Columna existe en v3 pero no en v4
- MISSING_IN_V3: Columna existe en v4 pero no en v3
- POSITION_MISMATCH: Columna en posición diferente
- TYPE_MISMATCH: Tipo de dato diferente (integer<->bigint permitido)

### Parity (datos)
- UNICIDAD_V4: Hay duplicados en v4
- ROWCOUNT: Diferente cantidad de filas
- ONLY_V3: Fila existe en v3 pero no en v4
- ONLY_V4: Fila existe en v4 pero no en v3
- NULL_MISMATCH: NULL en v3 vs NOT NULL en v4 (o viceversa)
- NUMERIC_DIFF: Diferencia numérica > 0.01

## Interpretación
- **0 errores** = PASS ✅
- **> 0 errores** = FAIL ❌ (script hace RAISE EXCEPTION)

## Archivos

| Script | Propósito |
|--------|-----------|
| parity_v3_lot_metrics.sql | Validar datos v3_lot_metrics vs v4_report.lot_metrics |
| parity_v3_labor_list.sql | Validar datos v3_labor_list vs v4_report.labor_list |
| parity_contract_v3_lot_metrics.sql | Validar schema (columnas) iguales |
| sanity_explain_v4_lot_metrics.sql | Verificar que la vista es ejecutable |
