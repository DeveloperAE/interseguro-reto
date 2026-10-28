import { useEffect, useState } from "react";
import { apiConfig, request } from "./api.js";
import { formatNumber, parseMatrix } from "./format.js";

const EXAMPLE = "1 2 3\n4 5 6";
const TOKEN_KEY = "matrix-token";

export function App() {
  const config = apiConfig();
  const [token, setToken] = useState(() => sessionStorage.getItem(TOKEN_KEY) || "");
  const [username, setUsername] = useState("demo");
  const [password, setPassword] = useState("");
  const [matrixText, setMatrixText] = useState(EXAMPLE);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [goResult, setGoResult] = useState(null);
  const [nodeStats, setNodeStats] = useState(null);

  useEffect(() => {
    if (token) {
      sessionStorage.setItem(TOKEN_KEY, token);
    } else {
      sessionStorage.removeItem(TOKEN_KEY);
    }
  }, [token]);

  async function onLogin(event) {
    event.preventDefault();
    setError("");
    setLoading(true);
    try {
      const session = await request(config.goApiUrl, "/v1/auth/login", {
        body: { username, password },
      });
      setToken(session.token);
      setPassword("");
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  async function onProcess(event) {
    event.preventDefault();
    setError("");
    setGoResult(null);
    setNodeStats(null);
    let matrix;
    try {
      matrix = parseMatrix(matrixText);
    } catch (err) {
      setError(err.message);
      return;
    }
    setLoading(true);
    try {
      const processed = await request(config.goApiUrl, "/v1/matrices/process", {
        token,
        body: { matrix },
      });
      setGoResult(processed);
      try {
        const direct = await request(config.nodeApiUrl, "/v1/statistics", {
          token,
          body: {
            matrices: [
              { name: "rotated", values: processed.rotated },
              { name: "q", values: processed.qr.q },
              { name: "r", values: processed.qr.r },
            ],
          },
        });
        setNodeStats(direct);
      } catch (err) {
        setError(`Go respondió, pero Node no pudo ser consultado desde el navegador: ${err.message}`);
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  function logout() {
    setToken("");
    setGoResult(null);
    setNodeStats(null);
    setError("");
  }

  return (
    <div className="page">
      <header className="top">
        <p className="brand">Interseguro · reto técnico</p>
        <h1>Procesamiento de matrices</h1>
        <p className="lead">
          Go recibe la matriz, la rota 90° en sentido horario y calcula su factorización QR.
          Después envía las matrices resultantes (Rotada, Q y R) a Node para calcular las estadísticas.
        </p>
      </header>

      {error ? (
        <p className="banner" role="alert">
          {error}
        </p>
      ) : null}

      {!token ? (
        <form className="card login" onSubmit={onLogin}>
          <h2>Ingreso</h2>
          <p>El mismo token sirve para las dos APIs. Revisión: usuario <strong>demo</strong>, contraseña <strong>demo123</strong>.</p>
          <label>
            Usuario
            <input value={username} onChange={(event) => setUsername(event.target.value)} autoComplete="username" required />
          </label>
          <label>
            Contraseña
            <input
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              autoComplete="current-password"
              required
            />
          </label>
          <button type="submit" disabled={loading}>
            {loading ? "Ingresando…" : "Ingresar"}
          </button>
        </form>
      ) : (
        <div className="workspace">
          <form className="card" onSubmit={onProcess}>
            <div className="card-head">
              <h2>Matriz de entrada</h2>
              <button type="button" className="ghost" onClick={logout}>
                Salir
              </button>
            </div>
            <p>Una fila por línea. Separa los números con espacios o comas.</p>
            <label>
              Valores
              <textarea
                value={matrixText}
                onChange={(event) => setMatrixText(event.target.value)}
                rows={4}
                spellCheck="false"
                required
              />
            </label>
            <div className="actions">
              <button type="submit" disabled={loading}>
                {loading ? "Procesando…" : "Procesar"}
              </button>
              <button type="button" className="ghost" onClick={() => setMatrixText(EXAMPLE)}>
                Usar ejemplo
              </button>
            </div>
          </form>

          {goResult ? (
            <section className="results" aria-live="polite">
              <article className="card">
                <h2>Rotación</h2>
                <p>Sentido horario. La matriz {shape(goResult.original)} pasa a {shape(goResult.rotated)}.</p>
                <div className="pair">
                  <MatrixTable title="Original" values={goResult.original} />
                  <MatrixTable title="Rotada" values={goResult.rotated} />
                </div>
              </article>

              <article className="card">
                <h2>Factorización QR</h2>
                <p>
                  De la matriz original: A ≈ Q·R. Error de reconstrucción:{" "}
                  <strong>{formatNumber(goResult.qr.reconstructionError)}</strong>
                </p>
                <div className="pair">
                  <MatrixTable title="Q" values={goResult.qr.q} />
                  <MatrixTable title="R" values={goResult.qr.r} />
                </div>
              </article>

              <article className="card">
                <h2>Estadísticas</h2>
                <p>Calculadas por Node sobre las matrices devueltas por Go (Rotada, Q y R). El navegador también consulta Node directamente.</p>
                <StatsTable
                  fromGo={goResult.statistics}
                  fromNode={nodeStats}
                />
                <MatrixDetailsTable details={goResult.statistics.matrices} />
              </article>
            </section>
          ) : (
            <p className="empty">El resultado aparece aquí después de procesar una matriz.</p>
          )}
        </div>
      )}
    </div>
  );
}

function shape(values) {
  return `${values.length}×${values[0]?.length ?? 0}`;
}

function MatrixTable({ title, values }) {
  return (
    <div className="table-wrap">
      <table>
        <caption>{title}</caption>
        <tbody>
          {values.map((row, i) => (
            <tr key={i}>
              {row.map((value, j) => (
                <td key={j}>{formatNumber(value)}</td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function StatsTable({ fromGo, fromNode }) {
  const rows = [
    ["Máximo", fromGo.max, fromNode?.max],
    ["Mínimo", fromGo.min, fromNode?.min],
    ["Promedio", fromGo.average, fromNode?.average],
    ["Suma", fromGo.sum, fromNode?.sum],
    ["¿Alguna es diagonal?", yesNo(fromGo.hasDiagonalMatrix), fromNode ? yesNo(fromNode.hasDiagonalMatrix) : "—"],
  ];
  return (
    <div className="table-wrap">
      <table className="stats">
        <caption>Comparación de las dos respuestas</caption>
        <thead>
          <tr>
            <th scope="col">Dato</th>
            <th scope="col">Incluido en la respuesta de Go</th>
            <th scope="col">Consulta directa a Node</th>
          </tr>
        </thead>
        <tbody>
          {rows.map(([label, goValue, nodeValue]) => (
            <tr key={label}>
              <th scope="row">{label}</th>
              <td>{typeof goValue === "number" ? formatNumber(goValue) : goValue}</td>
              <td>{typeof nodeValue === "number" ? formatNumber(nodeValue) : nodeValue}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function yesNo(value) {
  return value ? "Sí" : "No";
}

const MATRIX_LABELS = {
  rotated: "Rotada (90° horario)",
  q: "Q (ortogonal)",
  r: "R (triangular superior)",
};

function MatrixDetailsTable({ details }) {
  if (!Array.isArray(details) || details.length === 0) {
    return null;
  }
  return (
    <div className="table-wrap" style={{ marginTop: "16px" }}>
      <table className="stats">
        <caption>Detalle por matriz evaluada</caption>
        <thead>
          <tr>
            <th scope="col">Matriz</th>
            <th scope="col">Dimensiones</th>
            <th scope="col">¿Es diagonal?</th>
          </tr>
        </thead>
        <tbody>
          {details.map((item) => (
            <tr key={item.name}>
              <th scope="row">{MATRIX_LABELS[item.name] || item.name}</th>
              <td>
                {item.rows}×{item.cols}
              </td>
              <td>{yesNo(item.isDiagonal)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
