// Capa de Aplicación (Application Use Case Layer) siguiendo Arquitectura Hexagonal.
import { analyze } from "../domain/statistics.js";

/**
 * Crea el caso de uso de análisis estadístico de matrices inyectando el servicio de dominio.
 * @param {(matrices: Array<{name?: string, values: number[][]}>) => object} domainAnalyzer
 */
export function createAnalyzeMatricesUseCase(domainAnalyzer = analyze) {
  return {
    execute(matrices) {
      return domainAnalyzer(matrices);
    },
  };
}
