const { execSync } = require("node:child_process");

// #372: the kit HTML lives inside the Go module, not in web/templates, so
// Tailwind never sees its utility classes and a fresh app renders unstyled
// components. Resolve the module the same way Go does; a local replace
// (amarra-cais link) keeps working.
function shippedKitContent() {
  try {
    const out = execSync("go list -m -f {{.Dir}} github.com/puppe1990/amarra-cais", {
      encoding: "utf8",
      stdio: ["ignore", "pipe", "ignore"],
    });
    const dir = out.trim().replace(/\\/g, "/");
    if (!dir) {
      console.warn(
        "amarra-cais: shippedKitContent: empty module dir; kit-classes.txt still covers kit utilities. Run go mod download before npm run build (#391)."
      );
      return [];
    }
    return [dir + "/pkg/amarra/view/components/**/*.html"];
  } catch (err) {
    console.warn(
      "amarra-cais: shippedKitContent: go list -m failed; kit-classes.txt still covers kit utilities. Run go mod download before npm run build (#391).",
      err && err.message ? err.message : err
    );
    return [];
  }
}

/** @type {import("tailwindcss").Config} */
module.exports = {
  content: [
    "./web/templates/**/*.html",
    "./web/static/css/kit-classes.txt",
    ...shippedKitContent(),
  ],
  safelist: [
    "cais-password-wrap",
    "cais-password-toggle",
    "cais-chat-scroll-down",
    "cais-thinking",
    "cais-thinking-dots",
    "cais-select-search",
    "cais-select-search-native",
    "cais-select-search-trigger",
    "cais-select-search-panel",
    "cais-select-search-input",
    "cais-select-search-list",
    "cais-select-search-option",
    "cais-select-search-label",
    "cais-select-search-chevron",
    "is-selected",
    "is-highlighted",
    "is-hidden",
  ],
  theme: {
    extend: {
      colors: {
        /* CSS variables (input.css) so html.light can flip the palette (#258). */
        ink: "rgb(var(--amarra-ink) / <alpha-value>)",
        foam: "rgb(var(--amarra-foam) / <alpha-value>)",
        copper: "rgb(var(--amarra-copper) / <alpha-value>)",
        tide: "rgb(var(--amarra-tide) / <alpha-value>)",
        coral: "rgb(var(--amarra-coral) / <alpha-value>)",
        /* Toned semantic pair per theme — never stock red/green from Tailwind. */
        danger: "rgb(var(--amarra-danger) / <alpha-value>)",
        success: "rgb(var(--amarra-success) / <alpha-value>)",
      },
      fontFamily: {
        sans: ["Segoe UI", "Helvetica Neue", "system-ui", "sans-serif"],
        serif: [
          "Instrument Serif",
          "Iowan Old Style",
          "Palatino Linotype",
          "Palatino",
          "Georgia",
          "serif",
        ],
        display: [
          "Instrument Serif",
          "Iowan Old Style",
          "Palatino Linotype",
          "Palatino",
          "Georgia",
          "serif",
        ],
        mono: ["IBM Plex Mono", "ui-monospace", "SFMono-Regular", "Menlo", "monospace"],
      },
      boxShadow: {
        "2xs": "0 1px 2px 0 rgb(0 0 0 / 0.05)",
        xs: "0 1px 2px 0 rgb(0 0 0 / 0.05)",
      },
    },
  },
  plugins: [],
};
