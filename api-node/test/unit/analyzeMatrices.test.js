import assert from "node:assert/strict";
import test from "node:test";
import { createAnalyzeMatricesUseCase } from "../../src/usecases/analyzeMatrices.js";

test("createAnalyzeMatricesUseCase delega al analizador de dominio inyectado", () => {
  let calledWith = null;
  const fakeAnalyzer = (matrices) => {
    calledWith = matrices;
    return { max: 10, min: 1, average: 5.5, sum: 11, hasDiagonalMatrix: true, matrices: [] };
  };
  const useCase = createAnalyzeMatricesUseCase(fakeAnalyzer);
  const input = [{ name: "q", values: [[1, 0], [0, 10]] }];
  const output = useCase.execute(input);
  assert.deepEqual(calledWith, input);
  assert.equal(output.max, 10);
});

test("por defecto usa el analizador real del dominio", () => {
  const output = createAnalyzeMatricesUseCase().execute([{ values: [[1, 0], [0, 3]] }]);
  assert.equal(output.sum, 4);
  assert.equal(output.hasDiagonalMatrix, true);
});
