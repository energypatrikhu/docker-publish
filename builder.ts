// Build cli
await Bun.build({
  entrypoints: ["./index.ts"],
  outdir: "./dist",
  format: "esm",
  target: "bun",
  minify: true,
});

// Add shebang to the generated file
const distFilePath = "./dist/index.js";
const shebang = `#!/usr/bin/env bun\n`;
const distFileContent = await Bun.file(distFilePath).text();
const updatedContent = shebang + distFileContent;
await Bun.write(distFilePath, updatedContent);
