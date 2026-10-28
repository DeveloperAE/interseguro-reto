// Composition root: valida la configuración, arma la app y escucha.
import { createApp } from "./app.js";
import { loadConfig } from "./config/env.js";

let config;
try {
  config = loadConfig();
} catch (error) {
  console.error(error.message);
  process.exit(1);
}

const app = createApp(config);
app.listen(config.port, () => {
  console.log(`api-node escuchando en :${config.port}`);
});
