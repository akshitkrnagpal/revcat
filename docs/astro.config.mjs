// @ts-check
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

const commands = [
  { label: "apps", slug: "commands/apps" },
  { label: "audit-logs", slug: "commands/audit-logs" },
  { label: "auth", slug: "commands/auth" },
  { label: "charts", slug: "commands/charts" },
  { label: "doctor", slug: "commands/doctor" },
  { label: "entitlements", slug: "commands/entitlements" },
  { label: "invoices", slug: "commands/invoices" },
  { label: "metrics", slug: "commands/metrics" },
  { label: "offerings", slug: "commands/offerings" },
  { label: "packages", slug: "commands/packages" },
  { label: "paywalls", slug: "commands/paywalls" },
  { label: "products", slug: "commands/products" },
  { label: "projects", slug: "commands/projects" },
  { label: "publish", slug: "commands/publish" },
  { label: "purchases", slug: "commands/purchases" },
  { label: "subscribers", slug: "commands/subscribers" },
  { label: "subscriptions", slug: "commands/subscriptions" },
  { label: "virtual-currencies", slug: "commands/virtual-currencies" },
  { label: "webhooks", slug: "commands/webhooks" },
];

export default defineConfig({
  site: "https://revcat.vercel.app",
  integrations: [
    starlight({
      title: "revcat",
      description:
        "Manage RevenueCat entitlements, offerings, paywalls, customers, webhooks, and audit logs from the terminal. Single static binary, JSON-first when piped.",
      social: [
        {
          icon: "github",
          label: "GitHub",
          href: "https://github.com/akshitkrnagpal/revcat",
        },
      ],
      head: [
        // Social-friendly title + description (longer than the in-app title
        // for SEO + link previews; opengraph.xyz wants 50-60 / 110-160).
        { tag: "meta", attrs: { property: "og:title", content: "revcat - the RevenueCat CLI for terminal-first workflows" } },
        { tag: "meta", attrs: { name: "twitter:title", content: "revcat - the RevenueCat CLI for terminal-first workflows" } },
        { tag: "meta", attrs: { property: "og:description", content: "Manage RevenueCat entitlements, offerings, paywalls, customers, webhooks, and audit logs from the terminal. Single static binary, JSON-first when piped." } },
        { tag: "meta", attrs: { name: "twitter:description", content: "Manage RevenueCat entitlements, offerings, paywalls, customers, webhooks, and audit logs from the terminal. Single static binary, JSON-first when piped." } },
        { tag: "meta", attrs: { property: "og:image", content: "https://revcat.vercel.app/og.png" } },
        { tag: "meta", attrs: { property: "og:image:width", content: "1200" } },
        { tag: "meta", attrs: { property: "og:image:height", content: "630" } },
        { tag: "meta", attrs: { property: "og:image:alt", content: "revcat - the RevenueCat CLI" } },
        { tag: "meta", attrs: { name: "twitter:image", content: "https://revcat.vercel.app/og.png" } },
        { tag: "meta", attrs: { name: "twitter:image:alt", content: "revcat - the RevenueCat CLI" } },
      ],
      // Force dark by default; user can still toggle.
      customCss: ["./src/styles/custom.css"],
      sidebar: [
        {
          label: "Getting Started",
          items: [
            { label: "Installation", slug: "getting-started/installation" },
            { label: "Quickstart", slug: "getting-started/quickstart" },
          ],
        },
        {
          label: "Commands",
          items: [
            { label: "Overview", slug: "commands" },
            ...commands,
          ],
        },
        {
          label: "Guides",
          autogenerate: { directory: "guides" },
        },
        {
          label: "Reference",
          autogenerate: { directory: "reference" },
        },
      ],
    }),
  ],
});
