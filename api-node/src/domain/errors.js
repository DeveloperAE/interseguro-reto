// Error de entrada inválida del dominio. El adaptador HTTP lo traduce a 400.
export class ValidationError extends Error {
  constructor(message) {
    super(message);
    this.name = "ValidationError";
  }
}
