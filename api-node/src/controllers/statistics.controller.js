export function createStatisticsController(analyzeUseCase) {
  return {
    analyze(req, res, next) {
      try {
        return res.json(analyzeUseCase.execute(req.body?.matrices));
      } catch (error) {
        return next(error);
      }
    },
  };
}
