import assert from "node:assert/strict";
import test from "node:test";
import { ValidationError } from "../../src/domain/errors.js";
import { analyze, DIAGONAL_EPSILON, isDiagonal, MAX_DIMENSION } from "../../src/domain/statistics.js";

const sample = [
  {
    name: "rotated",
    values: [
      [4, 1],
      [5, 2],
      [6, 3],
    ],
  },
];

test("calcula máximo, mínimo, promedio y suma de la matriz de ejemplo", () => {
  const result = analyze(sample);
  assert.equal(result.max, 6);
  assert.equal(result.min, 1);
  assert.equal(result.sum, 21);
  assert.equal(result.average, 3.5);
  assert.equal(result.hasDiagonalMatrix, false);
  assert.deepEqual(result.matrices[0], {
    name: "rotated",
    isDiagonal: false,
    rows: 3,
    cols: 2,
  });
});

test("marca diagonal solo si alguna matriz lo es", () => {
  const result = analyze([
    { name: "rotated", values: [[4, 1], [5, 2], [6, 3]] },
    { name: "identity", values: [[1, 0], [0, 2]] },
  ]);
  assert.equal(result.hasDiagonalMatrix, true);
  assert.equal(result.matrices[0].isDiagonal, false);
  assert.equal(result.matrices[1].isDiagonal, true);
  assert.equal(result.max, 6);
  assert.equal(result.sum, 24);
});

test("una matriz no cuadrada no es diagonal", () => {
  assert.equal(isDiagonal([[1, 0, 0], [0, 2, 0]]), false);
});

test("un escalar es diagonal", () => {
  assert.equal(analyze([{ values: [[0]] }]).hasDiagonalMatrix, true);
});

test("tolera residuo numérico de punto flotante fuera de la diagonal", () => {
  assert.equal(DIAGONAL_EPSILON, 1e-9);
  assert.equal(isDiagonal([[1, 1e-15], [-1e-14, 2]]), true);
  assert.equal(isDiagonal([[1, 1e-4], [0, 2]]), false);
});

test("rechaza matrices irregulares, vacías, no numéricas y que excedan MAX_DIMENSION", () => {
  assert.throws(() => analyze([]), /al menos una matriz/);
  assert.throws(() => analyze([{ values: [[1, 2], [3]] }]), /rectangular/);
  assert.throws(() => analyze([{ values: [[Number.NaN]] }]), /finito/);
  assert.throws(() => analyze([{ values: [] }]), /al menos una fila/);
  assert.throws(() => analyze([{ values: [[]] }]), /al menos una columna/);

  const oversized = Array.from({ length: MAX_DIMENSION + 1 }, () => [1]);
  assert.throws(() => analyze([{ values: oversized }]), /tamaño máximo permitido/);
});

test("los errores de entrada son ValidationError del dominio", () => {
  assert.throws(() => analyze(undefined), ValidationError);
});
