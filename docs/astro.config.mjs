// @ts-check
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

const commands = [
  { label: "apps", slug: "commands/apps" },
  { label: "auth", slug: "commands/auth" },
  { label: "charts", slug: "commands/charts" },
  { label: "doctor", slug: "commands/doctor" },
  { label: "entitlements", slug: "commands/entitlements" },
  { label: "events", slug: "commands/events" },
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
        "The RevenueCat CLI. Run your RevenueCat project from the terminal instead of clicking through the dashboard.",
      social: [
        {
          icon: "github",
          label: "GitHub",
          href: "https://github.com/akshitkrnagpal/revcat",
        },
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
