import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.tsx";

import { initializeI18n } from "./i18n";

const rootElement = document.getElementById("root")!;
const root = createRoot(rootElement);

const renderLoading = () => {
  root.render(
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        height: "100vh",
        fontSize: "16px",
        color: "#666",
      }}
    >
      Loading...
    </div>
  );
};

const renderApp = () => {
  root.render(
    <StrictMode>
      <App />
    </StrictMode>
  );
};

const initApp = async () => {
  try {
    renderLoading();

    await initializeI18n();

    renderApp();
  } catch (error) {
    console.error("Failed to initialize application:", error);

    root.render(
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          height: "100vh",
          fontSize: "16px",
          color: "#ff4444",
          textAlign: "center",
          padding: "20px",
        }}
      >
        <div>
          <div>Failed to load application</div>
          <div style={{ fontSize: "14px", marginTop: "8px", opacity: 0.7 }}>
            Please refresh the page to try again
          </div>
        </div>
      </div>
    );
  }
};

initApp();
