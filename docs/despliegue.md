# Despliegue

Cómo repetir el blueprint en Render. Las URLs públicas vigentes están en el [README](../README.md) y no llevan barra final.

`render.yaml` crea tres web services Docker en el plan gratuito de Render: `matrix-api-node`, `matrix-api-go` y `matrix-web`.

1. Sube el repositorio a GitHub. `.env` está en `.gitignore` y no se sube; el secreto local no sale de tu máquina.
2. En Render: **New → Blueprint**, elige el repositorio y confirma `render.yaml`.
3. Render pide los valores marcados con `sync: false`:
   - `DEMO_PASSWORD` en `matrix-api-node`: `demo123`. Go lo toma de Node.
   - `CORS_ORIGIN` en las dos APIs: `*` por ahora.
   - `NODE_API_URL` en `matrix-api-go`, y `GO_API_URL` / `NODE_API_URL` en `matrix-web`: se pueden dejar vacías, todavía no se conocen las URLs.

   `JWT_SECRET` no se pide: Render lo genera en Node (256 bits en base64) y Go lo lee de ahí.
4. Cuando termine el primer despliegue, copia la URL pública de cada servicio (`https://<servicio>.onrender.com`) y en **Environment** completa:

   | Servicio | Variable | Valor |
   | --- | --- | --- |
   | `matrix-api-go` | `NODE_API_URL` | URL de `matrix-api-node` |
   | `matrix-web` | `GO_API_URL` | URL de `matrix-api-go` |
   | `matrix-web` | `NODE_API_URL` | URL de `matrix-api-node` |
   | las dos APIs | `CORS_ORIGIN` | URL de `matrix-web` |

   Las URLs no deben terminar en `/` (forma correcta: `https://<servicio>.onrender.com`). El frontend tolera una barra final, pero documéntalas sin ella.

   Guarda con redeploy. Go llama a Node por su URL pública: en el plan gratuito un web service no recibe tráfico de la red privada.

`TRUST_PROXY=true` está activo solo en Render, para que el límite de login use la IP de `X-Forwarded-For` y no la del proxy (ver [decisiones.md](decisiones.md)).

Si `NODE_API_URL` está vacía o termina en `/`, `POST /v1/matrices/process` responde `502`. `CORS_ORIGIN` debe ser la URL del frontend, sin barra final (hoy, `https://matrix-web-albq.onrender.com`). Si CORS está mal, el navegador falla y la llamada directa a las APIs sigue funcionando.

El plan gratuito duerme los servicios; el primer `GET /health` puede tardar cerca de un minuto.
