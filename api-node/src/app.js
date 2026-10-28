// Adaptador Primario HTTP (Driving Adapter) en Express:
// compone middlewares y rutas; el dominio y los casos de uso no dependen de Express.
import express from "express";
import cors from "cors";
import { LOGIN_MAX_ATTEMPTS, LOGIN_WINDOW_MS } from "./config/env.js";
import { errorHandler } from "./middlewares/errorHandler.js";
import { createRouter } from "./routes/index.js";
import { createAnalyzeMatricesUseCase } from "./usecases/analyzeMatrices.js";

export function createApp(config, { analyzeUseCase = createAnalyzeMatricesUseCase() } = {}) {
  const settings = {
    ...config,
    loginRateLimit: config.loginRateLimit ?? { windowMs: LOGIN_WINDOW_MS, limit: LOGIN_MAX_ATTEMPTS },
  };

  const app = express();
  app.disable("x-powered-by");
  if (settings.trustProxy) {
    // Un solo salto (el proxy de Render): req.ip pasa a ser la IP del cliente para el rate limit.
    app.set("trust proxy", 1);
  }
  app.use(express.json({ limit: "1mb" }));
  app.use((req, res, next) => {
    res.setHeader("X-Content-Type-Options", "nosniff");
    next();
  });
  app.use(
    cors({
      origin: settings.corsOrigin === "*" ? true : settings.corsOrigin,
      allowedHeaders: ["Content-Type", "Authorization"],
      methods: ["GET", "POST", "OPTIONS"],
    }),
  );

  app.use(createRouter(settings, { analyzeUseCase }));
  app.use(errorHandler);

  return app;
}
