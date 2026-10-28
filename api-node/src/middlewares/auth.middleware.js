import { verifyToken } from "../services/token.service.js";

export function requireAuth(secret) {
  return (req, res, next) => {
    const header = req.get("authorization") ?? "";
    const [scheme, token] = header.split(" ");
    if (scheme?.toLowerCase() !== "bearer" || !token) {
      return res.status(401).json({ error: "se requiere un token Bearer" });
    }
    try {
      verifyToken(secret, token);
      return next();
    } catch {
      return res.status(401).json({ error: "token inválido o vencido" });
    }
  };
}
