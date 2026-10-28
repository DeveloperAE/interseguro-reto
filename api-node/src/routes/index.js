import { Router } from "express";
import { createAuthRouter } from "./auth.routes.js";
import { createHealthRouter } from "./health.routes.js";
import { createStatisticsRouter } from "./statistics.routes.js";

export function createRouter(config, { analyzeUseCase }) {
  const router = Router();
  router.use(createHealthRouter());
  router.use("/v1/auth", createAuthRouter(config));
  router.use("/v1/statistics", createStatisticsRouter(config, analyzeUseCase));
  return router;
}
