const { readdirSync } = require("node:fs")

const featureNames = readdirSync("src/features", { withFileTypes: true })
  .filter((entry) => entry.isDirectory())
  .map((entry) => entry.name)

module.exports = {
  forbidden: [
    {
      name: "bootstrap-only-imports-app",
      severity: "error",
      from: { path: "^src/main\\.tsx$" },
      to: { path: "^src/", pathNot: "^src/app/" },
    },
    {
      name: "no-circular-dependencies",
      severity: "error",
      from: {},
      to: { circular: true },
    },
    {
      name: "no-unresolved-imports",
      severity: "error",
      from: {},
      to: { couldNotResolve: true },
    },
    {
      name: "shared-does-not-import-business",
      severity: "error",
      from: { path: "^src/shared/" },
      to: { path: "^src/(app|features)/" },
    },
    {
      name: "features-do-not-import-app",
      severity: "error",
      from: { path: "^src/features/" },
      to: { path: "^src/app/" },
    },
    {
      name: "application-uses-public-feature-entries",
      severity: "error",
      from: { path: "^(src/app/|src/main\\.tsx$|\\.storybook/)" },
      to: { path: "^src/features/[^/]+/(?!index\\.ts$)" },
    },
    {
      name: "feature-model-does-not-import-ui",
      severity: "error",
      from: { path: "^src/features/[^/]+/model/" },
      to: { path: "^src/(features/[^/]+/ui/|shared/ui/)" },
    },
    {
      name: "feature-api-only-uses-data-and-infrastructure",
      severity: "error",
      from: { path: "^src/features/[^/]+/api/" },
      to: { path: "^(src/features/[^/]+/(model|ui)/|src/shared/(ui|hooks)/|node_modules/(react|react-dom|react-router|@tanstack/react-query)/)" },
    },
    {
      name: "feature-ui-uses-model",
      severity: "error",
      from: { path: "^src/features/[^/]+/ui/" },
      to: { path: "^(src/features/[^/]+/api/|node_modules/(@tanstack/react-query|axios)/|src/shared/api/)" },
    },
    {
      name: "feature-types-are-leaves",
      severity: "error",
      from: { path: "^src/features/[^/]+/types\\.ts$" },
      to: { path: "^src/" },
    },
    {
      name: "shared-foundations-do-not-import-presentation",
      severity: "error",
      from: { path: "^src/shared/(api|lib|config|i18n)/" },
      to: { path: "^src/shared/(ui|hooks)/" },
    },
    {
      name: "shared-hooks-do-not-import-ui",
      severity: "error",
      from: { path: "^src/shared/hooks/" },
      to: { path: "^src/shared/ui/" },
    },
    {
      name: "no-legacy-directories",
      severity: "error",
      from: {},
      to: { path: "^src/(api|auth|components|config|hooks|i18n|lib|merchant-console|pages)/" },
    },
    ...featureNames.flatMap((feature) => [
      {
        name: `${feature}-does-not-import-other-features`,
        severity: "error",
        from: { path: `^src/features/${feature}/` },
        to: { path: "^src/features/", pathNot: `^src/features/${feature}/` },
      },
      {
        name: `${feature}-internals-do-not-import-public-entry`,
        severity: "error",
        from: { path: `^src/features/${feature}/`, pathNot: `^src/features/${feature}/index\\.ts$` },
        to: { path: `^src/features/${feature}/index\\.ts$` },
      },
    ]),
  ],
  options: {
    doNotFollow: { path: "node_modules" },
    parser: "swc",
    tsConfig: { fileName: "tsconfig.app.json" },
    enhancedResolveOptions: {
      exportsFields: ["exports"],
      extensions: [".ts", ".tsx", ".js", ".jsx", ".json", ".css"],
      conditionNames: ["types", "import", "browser", "default"],
    },
  },
}
