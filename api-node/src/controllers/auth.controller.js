import { credentialsMatch, signToken } from "../services/token.service.js";

export function createAuthController(config) {
  return {
    login(req, res) {
      const { username, password } = req.body ?? {};
      if (!credentialsMatch(username, password, config.demoUsername, config.demoPassword)) {
        return res.status(401).json({ error: "credenciales inválidas" });
      }
      const token = signToken(config.jwtSecret, username, config.tokenTtlSeconds);
      return res.json({ token, expiresIn: config.tokenTtlSeconds });
    },
  };
}
