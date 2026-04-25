import { readFile, writeFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";
import satori from "satori";
import { Resvg } from "@resvg/resvg-js";

const here = dirname(fileURLToPath(import.meta.url));
const fontsDir = join(here, "fonts");

const [sansBold, sansMedium, monoBold, monoMedium] = await Promise.all([
  readFile(join(fontsDir, "DMSans-Bold.ttf")),
  readFile(join(fontsDir, "DMSans-Medium.ttf")),
  readFile(join(fontsDir, "JetBrainsMono-Bold.ttf")),
  readFile(join(fontsDir, "JetBrainsMono-Medium.ttf")),
]);

const PAPER = "#0a0a0b";
const INK = "#e8e6e3";
const MUTED = "#8a8a8a";
const ACCENT = "#7c8aff"; // soft RC-ish blue
const ACCENT2 = "#26d07c"; // green like a successful command

const el = (type, props) => ({ type, props });

const tree = el("div", {
  style: {
    width: "1200px",
    height: "630px",
    background: PAPER,
    color: INK,
    fontFamily: "DM Sans",
    position: "relative",
    display: "flex",
    flexDirection: "column",
    overflow: "hidden",
  },
  children: [
    // Soft accent blobs
    el("div", {
      style: {
        position: "absolute",
        top: "-180px",
        right: "-160px",
        width: "720px",
        height: "720px",
        borderRadius: "720px",
        background: ACCENT,
        opacity: 0.15,
        filter: "blur(120px)",
      },
    }),
    el("div", {
      style: {
        position: "absolute",
        bottom: "-160px",
        left: "-120px",
        width: "560px",
        height: "560px",
        borderRadius: "560px",
        background: ACCENT2,
        opacity: 0.08,
        filter: "blur(120px)",
      },
    }),
    // Header row
    el("div", {
      style: {
        position: "absolute",
        top: "56px",
        left: "64px",
        right: "64px",
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        color: MUTED,
        fontSize: "20px",
        letterSpacing: "6px",
        textTransform: "uppercase",
      },
      children: [
        el("span", { children: "revcat" }),
        el("span", { style: { color: INK, letterSpacing: "2px" }, children: "revcat.vercel.app" }),
      ],
    }),
    // Body
    el("div", {
      style: {
        flex: 1,
        display: "flex",
        flexDirection: "column",
        alignItems: "flex-start",
        justifyContent: "center",
        padding: "0 96px",
      },
      children: [
        el("div", {
          style: {
            fontSize: "92px",
            fontWeight: 700,
            lineHeight: 1.05,
            letterSpacing: "-2px",
            color: INK,
            maxWidth: "1000px",
          },
          children: "The RevenueCat CLI.",
        }),
        el("div", {
          style: {
            marginTop: "24px",
            fontSize: "32px",
            fontWeight: 500,
            color: MUTED,
            lineHeight: 1.35,
            maxWidth: "1000px",
          },
          children:
            "Run your project from the terminal instead of clicking through the dashboard.",
        }),
        // Code block (simple single-line for satori compatibility)
        el("div", {
          style: {
            marginTop: "44px",
            padding: "20px 28px",
            background: "#15151a",
            border: `1.5px solid #2a2a33`,
            borderRadius: "14px",
            fontFamily: "JetBrains Mono",
            fontSize: "26px",
            color: ACCENT2,
            display: "flex",
          },
          children: "$ revcat subscribers info app_user_123",
        }),
      ],
    }),
    // Bottom accent line
    el("div", {
      style: {
        position: "absolute",
        bottom: 0,
        left: 0,
        right: 0,
        height: "4px",
        background: `linear-gradient(90deg, transparent 0%, ${ACCENT} 50%, transparent 100%)`,
      },
    }),
  ],
});

const svg = await satori(tree, {
  width: 1200,
  height: 630,
  fonts: [
    { name: "DM Sans", data: sansBold, weight: 700, style: "normal" },
    { name: "DM Sans", data: sansMedium, weight: 500, style: "normal" },
    { name: "JetBrains Mono", data: monoBold, weight: 700, style: "normal" },
    { name: "JetBrains Mono", data: monoMedium, weight: 500, style: "normal" },
  ],
});

const png = new Resvg(svg, { fitTo: { mode: "width", value: 1200 } })
  .render()
  .asPng();

const outPath = join(here, "..", "public", "og.png");
await writeFile(outPath, png);
console.log(`wrote ${outPath} (${png.length} bytes)`);
