// Capa de Dominio (Domain Layer): validación de matrices, detección de diagonal y cálculo estadístico.
import { ValidationError } from "./errors.js";

export const MAX_DIMENSION = 250;
export const DIAGONAL_EPSILON = 1e-9;

function isFiniteNumber(value) {
  return typeof value === "number" && Number.isFinite(value);
}

function assertRectangular(values) {
  if (!Array.isArray(values) || values.length === 0) {
    throw new ValidationError("cada matriz debe tener al menos una fila");
  }
  const cols = values[0]?.length ?? 0;
  if (!Array.isArray(values[0]) || cols === 0) {
    throw new ValidationError("cada matriz debe tener al menos una columna");
  }
  if (values.length > MAX_DIMENSION || cols > MAX_DIMENSION) {
    throw new ValidationError(
      `la matriz excede el tamaño máximo permitido (${MAX_DIMENSION}x${MAX_DIMENSION})`,
    );
  }
  for (let i = 0; i < values.length; i += 1) {
    const row = values[i];
    if (!Array.isArray(row) || row.length !== cols) {
      throw new ValidationError(`la matriz no es rectangular en la fila ${i}`);
    }
    for (let j = 0; j < row.length; j += 1) {
      if (!isFiniteNumber(row[j])) {
        throw new ValidationError(`el valor en [${i}][${j}] no es un número finito`);
      }
    }
  }
}

export function isDiagonal(values) {
  const rows = values.length;
  const cols = values[0].length;
  if (rows !== cols) {
    return false;
  }
  for (let i = 0; i < rows; i += 1) {
    for (let j = 0; j < cols; j += 1) {
      if (i !== j && Math.abs(values[i][j]) > DIAGONAL_EPSILON) {
        return false;
      }
    }
  }
  return true;
}

// analyze recorre todas las matrices una sola vez.
// El máximo, el mínimo, el promedio y la suma son globales.
// hasDiagonalMatrix es verdadero si alguna matriz es diagonal.
export function analyze(matrices) {
  if (!Array.isArray(matrices) || matrices.length === 0) {
    throw new ValidationError("se requiere al menos una matriz");
  }

  let max = -Infinity;
  let min = Infinity;
  let sum = 0;
  let count = 0;
  let hasDiagonalMatrix = false;
  const details = [];

  for (const matrix of matrices) {
    const values = matrix?.values;
    assertRectangular(values);
    const diagonal = isDiagonal(values);
    hasDiagonalMatrix = hasDiagonalMatrix || diagonal;
    for (const row of values) {
      for (const value of row) {
        if (value > max) max = value;
        if (value < min) min = value;
        sum += value;
        count += 1;
      }
    }
    details.push({
      name: typeof matrix.name === "string" && matrix.name !== "" ? matrix.name : "matrix",
      isDiagonal: diagonal,
      rows: values.length,
      cols: values[0].length,
    });
  }

  return {
    max,
    min,
    average: sum / count,
    sum,
    hasDiagonalMatrix,
    matrices: details,
  };
}
