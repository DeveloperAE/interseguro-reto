export function parseMatrix(text) {
  const lines = text
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean);
  if (lines.length === 0) {
    throw new Error("Escribe al menos una fila.");
  }
  const matrix = lines.map((line, rowIndex) => {
    const parts = line.split(/[\s,;]+/).filter(Boolean);
    if (parts.length === 0) {
      throw new Error(`La fila ${rowIndex + 1} está vacía.`);
    }
    return parts.map((part, colIndex) => {
      const value = Number(part);
      if (!Number.isFinite(value)) {
        throw new Error(`"${part}" no es un número (fila ${rowIndex + 1}, columna ${colIndex + 1}).`);
      }
      return value;
    });
  });
  const columns = matrix[0].length;
  for (let i = 1; i < matrix.length; i += 1) {
    if (matrix[i].length !== columns) {
      const found = matrix[i].length === 1 ? "1 valor" : `${matrix[i].length} valores`;
      const expected = columns === 1 ? "1 valor" : `${columns} valores`;
      throw new Error(`La fila ${i + 1} tiene ${found} y la primera tiene ${expected}.`);
    }
  }
  return matrix;
}

export function formatNumber(value) {
  if (!Number.isFinite(value)) {
    return "—";
  }
  if (Number.isInteger(value)) {
    return String(value);
  }
  const abs = Math.abs(value);
  if (abs !== 0 && (abs < 1e-4 || abs >= 1e6)) {
    return value.toExponential(3);
  }
  return new Intl.NumberFormat("es-PE", { maximumFractionDigits: 6 }).format(value);
}
