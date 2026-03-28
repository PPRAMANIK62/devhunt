import "@fontsource-variable/ibm-plex-sans";
import "@fontsource-variable/playfair-display";
import "@fontsource/ibm-plex-mono";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { AuthProvider } from "@/context/auth-context";
import "./index.css";
import App from "./app.tsx";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <BrowserRouter>
      <AuthProvider>
        <App />
      </AuthProvider>
    </BrowserRouter>
  </StrictMode>,
);
