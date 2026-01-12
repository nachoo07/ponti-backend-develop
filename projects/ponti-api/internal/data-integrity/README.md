# Data Integrity - Módulo de Control de Coherencia

## ⚠️  ADVERTENCIA CRÍTICA - NO MODIFICAR SIN AUTORIZACIÓN EXPLÍCITA ⚠️

**ESTOS CÁLCULOS SON CRÍTICOS Y NO DEBEN ALTERARSE A MENOS QUE SE RECIBA UNA ORDEN DIRECTA Y CLARA DEL USUARIO.**

### REGLAS INVIOLABLES:
- **NUNCA** modificar los cálculos LEFT/RIGHT sin autorización explícita
- **NUNCA** cambiar las tolerancias sin autorización explícita  
- **NUNCA** alterar la lógica de los 14 controles sin autorización explícita
- **NUNCA** usar ROUND() en cálculos internos (solo en DTOs de salida)
- **SIEMPRE** mantener precisión completa en cálculos SQL y Go

**Si necesitas modificar algo, DEBES pedir autorización explícita primero.**

---

## 📋 Descripción

Módulo para validar la coherencia de datos entre diferentes módulos del sistema mediante **14 controles automatizados** que garantizan la integridad de la información en toda la aplicación.

**Cada control calcula DOS valores de forma INDEPENDIENTE**:
- **LEFT** (Izquierda): Valor desde el **origen** → **Es el CORRECTO** (fuente de verdad)
- **RIGHT** (Derecha): Valor desde el **destino** → **Se valida contra LEFT**

---

## 🎯 Endpoint

```
GET /api/v1/data-integrity/costs-check?project_id=123
```

### Headers requeridos:
- `X-API-KEY`: abc123secreta
- `X-USER-ID`: 123

### Query Parameters:
- `project_id` (opcional): ID del proyecto a validar. Si no se envía, valida todos los proyectos.

---

## ✅ LOS 14 CONTROLES

### GRUPO 1: Costos Directos Ejecutados (Órdenes → varios destinos)

#### **Control 1**: Órdenes de trabajo → Dashboard
- **LEFT (Origen)**: `∑(workorders RAW)` - Suma directa de tablas workorders + workorder_items
- **RIGHT (Destino)**: `Dashboard.CostosDirectos` - Función SSOT
- **Tolerancia**: 0 (debe ser exacto)

#### **Control 2**: Órdenes de trabajo → Lotes
- **LEFT (Origen)**: `∑(workorders RAW)` - Suma directa de tablas
- **RIGHT (Destino)**: `∑(cost_usd_per_ha × sowed_area_ha)` - Vista v3_lot_list
- **Tolerancia**: 0 (debe ser exacto)

#### **Control 3**: Órdenes de trabajo → Informe de Resultado por campo
- **LEFT (Origen)**: `∑(workorders RAW)` - Suma directa de tablas
- **RIGHT (Destino)**: `∑(direct_cost_usd)` agrupado por field+crop
- **Tolerancia**: 0 (debe ser exacto)

#### **Control 4**: Órdenes de trabajo → Informe de Resultado Generales
- **LEFT (Origen)**: `∑(workorders RAW)` - Suma directa de tablas
- **RIGHT (Destino)**: Total general del informe (primera fila)
- **Tolerancia**: 0 (debe ser exacto)

---

### GRUPO 2: Invertidos (Labores + Insumos)

#### **Control 5**: Labores + Insumos → Dashboard
- **LEFT (Origen)**: `Labores + Semillas + Agroquímicos` - Suma de componentes
- **RIGHT (Destino)**: `Dashboard.Invertidos` - Total consolidado
- **Tolerancia**: ±1

#### **Control 6**: Labores → Informe de Aportes
- **LEFT (Origen)**: `Dashboard.LaboresInvertidos` - Summary
- **RIGHT (Destino)**: `∑(Labores Generales + Siembra + Riego)` - Sin cosecha
- **Tolerancia**: ±1

#### **Control 7**: Insumos → Informe de Aportes
- **LEFT (Origen)**: `Semillas + Agroquímicos` - Dashboard summary
- **RIGHT (Destino)**: `∑(Semilla + Agroquímicos)` - Informe de aportes
- **Tolerancia**: ±1

---

### GRUPO 3: Aportes desde Lotes

#### **Control 8**: Lotes → Informe de Aportes (Adm. Proyecto)
- **LEFT (Origen)**: `∑(admin_cost × sowed_area)` - Vista v3_lot_list
- **RIGHT (Destino)**: Categoría "Administración y Estructura" - Informe aportes
- **Tolerancia**: ±1

#### **Control 9**: Lotes → Informe de Aportes (Arriendo Fijo)
- **LEFT (Origen)**: `∑(rent_per_ha × sowed_area)` - Vista v3_lot_list
- **RIGHT (Destino)**: Categoría "Arriendo Capitalizable" - Informe aportes
- **Tolerancia**: ±1

---

### GRUPO 4: Ingreso Neto

#### **Control 10**: Lotes → Resumen de resultados
- **LEFT (Origen)**: `∑(income_net_per_ha × sowed_area)` - Vista v3_lot_list
- **RIGHT (Destino)**: `∑(net_income_usd)` - Resumen de resultados
- **Tolerancia**: ±1

---

### GRUPO 5: Resultado Operativo

#### **Control 11**: Lotes → Informe por cultivo
- **LEFT (Origen)**: `∑(operating_result_per_ha × sowed_area)` - Vista v3_lot_list
- **RIGHT (Destino)**: `∑(operating_result_usd_ha × sown_area_ha)` - Por cultivo
- **Tolerancia**: ±1

#### **Control 12**: Lotes → Informe de Resultado Generales
- **LEFT (Origen)**: `∑(operating_result_per_ha × sowed_area)` - Vista v3_lot_list
- **RIGHT (Destino)**: Total general (primera fila)
- **Tolerancia**: ±1

#### **Control 13**: Lotes → Dashboard
- **LEFT (Origen)**: `∑(operating_result_per_ha × sowed_area)` - Vista v3_lot_list
- **RIGHT (Destino)**: `Metrics.OperatingResult.ResultUSD` - Dashboard card
- **Tolerancia**: ±1

---

### GRUPO 6: Stock

#### **Control 14**: Stock → Dashboard
- **LEFT (Origen)**: `(Invertido - Ejecutado)` - Cálculo manual
- **RIGHT (Destino)**: `Summary.StockUSD` - Dashboard
- **Tolerancia**: ±1

---

## 📤 Respuesta

```json
{
  "checks": [
    {
      "control_number": 1,
      "source_module": "Órdenes de trabajo",
      "data_to_verify": "Costos directos ejecutados",
      "target_module": "Dashboard",
      "control_rule": "Dashboard.CostosDirectos = ∑(Ordenes.costo_total)",
      "left_calculation": "∑(workorders.effective_area × labors.price + workorder_items.total_used × supplies.price)",
      "left_value": "18454.39",
      "left_source": "Tabla workorders RAW",
      "right_calculation": "v3_dashboard_ssot.direct_costs_total_for_project()",
      "right_value": "18455.17",
      "right_source": "Vista v3_dashboard_management_balance",
      "difference": "-0.78",
      "status": "OK",
      "tolerance": "1.00"
    }
    // ... 13 controles más
  ]
}
```

---

## 🔧 Prueba con curl

```bash
curl -X GET "http://localhost:8080/api/v1/data-integrity/costs-check?project_id=11" \
  -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123"
```

---

## 📊 Estados posibles

- **OK**: La diferencia está dentro de la tolerancia
- **ERROR**: La diferencia excede la tolerancia

---

## 🏗️ Arquitectura

```
internal/data_integrity/
├── usecases/
│   └── domain/
│       └── types.go          # Tipos de dominio (IntegrityReport, IntegrityCheck con LEFT/RIGHT)
├── handler/
│   └── dto/
│       └── integrity_check.go # DTOs de respuesta con LEFT/RIGHT
├── usecases.go               # Lógica de negocio (14 controles con cálculos independientes)
└── handler.go                # HTTP handler
```

---

## 🔄 Flujo de validación

1. Recibir `project_id` opcional
2. Ejecutar **14 controles** en secuencia
3. Cada control:
   - **LEFT**: Calcula valor desde origen (fuente de verdad)
   - **RIGHT**: Calcula valor desde destino (a validar)
   - Calcula diferencia: `LEFT - RIGHT`
   - Aplica tolerancia
   - Determina estado (OK/ERROR)
4. Retornar lista de 14 controles con LEFT y RIGHT explícitos

---

## 📝 Notas importantes

- El scope es **por proyecto**: todas las validaciones se filtran por `project_id`
- **LEFT (origen) es SIEMPRE el valor correcto** - Es la fuente de verdad
- **RIGHT (destino) se valida contra LEFT**
- **Cálculos INDEPENDIENTES**: LEFT y RIGHT usan diferentes queries/vistas/tablas
- **No hay validación circular**: LEFT no usa las mismas funciones SSOT que RIGHT
- Los cálculos se hacen en **tiempo real** usando los repositorios existentes
- Cada control es **independiente** y se ejecuta en secuencia
- **Control 1-4**: LEFT usa datos RAW de workorders para evitar circularidad
- **Tolerancias**:
  - Controles 1-4: Tolerancia = 0 (deben ser exactos)
  - Controles 5-14: Tolerancia = ±1 USD para absorber diferencias de redondeo

---

## ✅ Estado de Implementación

**Total: 14/14 controles implementados** 🎉

### Controles por grupo:
- **Grupo 1 - Costos Directos**: 4/4 ✓
- **Grupo 2 - Invertidos**: 3/3 ✓
- **Grupo 3 - Aportes Lotes**: 2/2 ✓
- **Grupo 4 - Ingreso Neto**: 1/1 ✓
- **Grupo 5 - Resultado Operativo**: 3/3 ✓
- **Grupo 6 - Stock**: 1/1 ✓

---

## 📦 Módulos involucrados

- `workorder` - Órdenes de trabajo (costos directos ejecutados RAW)
- `dashboard` - Dashboard (balance de gestión, invertidos, stock)
- `lot` - Lotes (costos por ha, ingreso neto, resultado operativo)
- `report` - Informes (field-crop, summary-results, investor-contribution)
- `stock` - Stock (insumos en stock)

---

## 🔍 Ejemplo de uso

### Request:
```bash
GET /api/v1/data-integrity/costs-check?project_id=11
```

### Response:
```json
{
  "checks": [
    {
      "control_number": 2,
      "source_module": "Órdenes de trabajo",
      "data_to_verify": "Costos directos ejecutados",
      "target_module": "Lotes",
      "control_rule": "∑(Costo_directo_ha_lote × Superficie_lote) = ∑(Órdenes.costo_total)",
      "left_calculation": "∑(workorders RAW)",
      "left_value": "18454.39",
      "left_source": "Tabla workorders RAW",
      "right_calculation": "∑(cost_usd_per_ha × sowed_area_ha)",
      "right_value": "18454.39",
      "right_source": "Vista v3_lot_list",
      "difference": "0.00",
      "status": "OK",
      "tolerance": "0.00"
    }
  ]
}
```

---

## 🚨 Manejo de errores

Si algún control falla al ejecutarse (error de base de datos, módulo no disponible, etc.), el endpoint retornará un error 500 con el detalle del control que falló.

Ejemplo:
```json
{
  "error": "control 3 failed: failed to get field crop metrics: database error"
}
```

---

## 🎯 Ventajas de LEFT/RIGHT

1. **Transparencia total**: Se ve exactamente qué se está comparando
2. **No circular**: LEFT y RIGHT usan diferentes fuentes de datos
3. **Debugging fácil**: Si hay error, se ve cuál cálculo está mal
4. **Auditoría**: Cada valor tiene su fuente documentada
5. **Validación real**: LEFT (origen) valida RIGHT (destino)

---

## 📚 Referencias

- CSV de referencia: `Control de Integral.csv`
- Todos los controles están marcados con ✅ en el CSV original
- LEFT es siempre el valor de la columna "Pantalla origen"
- RIGHT es siempre el valor de la columna "Pantalla destino"
- Controles 1-4 tienen tolerancia 0 (deben ser exactos)
- Controles 5-14 tienen tolerancia ±1 USD
