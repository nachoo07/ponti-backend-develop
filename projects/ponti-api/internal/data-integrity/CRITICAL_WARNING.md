# ⚠️ ADVERTENCIA CRÍTICA - MÓDULO DATA-INTEGRITY ⚠️

## 🚨 REGLAS INVIOLABLES - NO MODIFICAR SIN AUTORIZACIÓN EXPLÍCITA 🚨

**ESTE MÓDULO CONTIENE CÁLCULOS CRÍTICOS DE INTEGRIDAD DE DATOS QUE NO DEBEN ALTERARSE A MENOS QUE SE RECIBA UNA ORDEN DIRECTA Y CLARA DEL USUARIO.**

### 📋 REGLAS INVIOLABLES:

1. **NUNCA** modificar los cálculos LEFT/RIGHT sin autorización explícita
2. **NUNCA** cambiar las tolerancias sin autorización explícita  
3. **NUNCA** alterar la lógica de los 14 controles sin autorización explícita
4. **NUNCA** usar ROUND() en cálculos internos (solo en DTOs de salida)
5. **SIEMPRE** mantener precisión completa en cálculos SQL y Go

### 🎯 CONTROLES CRÍTICOS:

- **Controles 1-4**: Tolerancia = 0 (deben ser exactos)
- **Controles 5-14**: Tolerancia = ±1 USD
- **LEFT**: Siempre el valor correcto (origen/fuente de verdad)
- **RIGHT**: Se valida contra LEFT

### ⚡ ACCIÓN REQUERIDA:

**Si necesitas modificar algo en este módulo, DEBES pedir autorización explícita primero.**

### 📝 ARCHIVOS PROTEGIDOS:

- `usecases.go` - Contiene los 14 controles críticos
- `usecases/domain/types.go` - Tipos de dominio críticos
- `handler.go` - Handler HTTP crítico
- `handler/dto/integrity_check.go` - DTOs críticos
- `README.md` - Documentación crítica

### 🔒 ESTADO:

**MÓDULO PROTEGIDO - NO MODIFICAR SIN AUTORIZACIÓN EXPLÍCITA**

---
*Última actualización: 2025-01-27*
*Creado por: Sistema de Protección de Integridad de Datos*
