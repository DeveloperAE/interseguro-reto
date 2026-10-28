import { Router } from "express";
import { createStatisticsController } from "../controllers/statistics.controller.js";
import { requireAuth } from "../middlewares/auth.middleware.js";

export function createStatisticsRouter(config, analyzeUseCase) {
  const router = Router();
  const controller = createStatisticsController(analyzeUseCase);
  router.post("/", requireAuth(config.jwtSecret), controller.analyze);
  return router;
}
