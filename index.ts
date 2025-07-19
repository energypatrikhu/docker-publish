import cliSelect from "cli-select";
import fs from "fs";
import path from "path";

const workingDirectory = process.argv[2] || process.cwd();
const dockerPublishFilePath = path.join(workingDirectory, ".docker-publish");

if (!fs.existsSync(dockerPublishFilePath)) {
  console.log("Initializing .docker-publish file...");

  console.log("Enter the Docker image name (e.g., energyhun24/bandwidth-hero-proxy):");
  const dockerImageName = await Bun.stdin.text();
  if (!dockerImageName) {
    console.error("Docker image name is required.");
    process.exit(1);
  }

  console.log("Enter the initial version (e.g., 1.0.0):");
  const initialVersion = await Bun.stdin.text();
  if (!initialVersion) {
    console.error("Initial version is required.");
    process.exit(1);
  }

  const dockerPublishJson = {
    dockerImageName: dockerImageName.trim(),
    version: initialVersion.trim(),
  };

  await Bun.write(dockerPublishFilePath, JSON.stringify(dockerPublishJson, null, "\t"));
  console.log(".docker-publish file created successfully.");
}

const dockerPublishJson = await Bun.file(dockerPublishFilePath).json();
const dockerImageName = dockerPublishJson.dockerImageName as string;
const version = dockerPublishJson.version as string;

const versionUpdateValues = {
  Current: version,
  Patch: updateVersion(version, "Patch"),
  Minor: updateVersion(version, "Minor"),
  Major: updateVersion(version, "Major"),
};
const versionUpdateValueKeys = Object.keys(versionUpdateValues) as (keyof typeof versionUpdateValues)[];

console.log("Select the version update type:");
const versionUpdate = await cliSelect({
  values: versionUpdateValueKeys,
  valueRenderer: (value: keyof typeof versionUpdateValues) => {
    return `${value} (${versionUpdateValues[value]})`;
  },
  cleanup: true,
});

const versionUpdateValue = versionUpdateValues[versionUpdate.value];

console.log(`You selected: ${versionUpdateValue}, current version: ${version}`);
console.log("Do you want to build, publish the image and update the version in the .version file?");
const buildAndPublishConfirm = await cliSelect({
  values: ["Yes", "No"],
  valueRenderer: (value) => {
    return value;
  },
  cleanup: true,
});

console.log("Do you want to push the changes to GIT?");
const pushChangesConfirm = await cliSelect({
  values: ["Yes", "No"],
  valueRenderer: (value) => {
    return value;
  },
  cleanup: true,
});

if (buildAndPublishConfirm.value === "Yes") {
  // Update the version in package.json
  dockerPublishJson.version = versionUpdateValue;
  await Bun.write(dockerPublishFilePath, JSON.stringify(dockerPublishJson, null, "\t"));
  console.log(`Version updated to: ${versionUpdateValue}`);

  // Build the Docker image
  await Bun.$`docker build -t ${dockerImageName}:${versionUpdateValue} -t ${dockerImageName}:latest ${workingDirectory}`;

  // Publish the image to Docker Hub
  await Bun.$`docker push ${dockerImageName}:${versionUpdateValue}`;
  await Bun.$`docker push ${dockerImageName}:latest`;

  if (pushChangesConfirm.value === "Yes" && versionUpdateValue !== version) {
    // Push changes to GIT
    await Bun.$`git add .docker-publish`;
    await Bun.$`git commit -m "chore: update version to ${versionUpdateValue}"`;
    await Bun.$`git push`;
    await Bun.$`git checkout main`;
    console.log("Changes pushed to GIT.");
  }
} else {
  console.log("Aborted.");
}

/* Utility functions */

function updateVersion(version: string, type: string) {
  const versionParts = version.split(".").map(Number);

  switch (type) {
    case "Major":
      versionParts[0]++;
      versionParts[1] = 0;
      versionParts[2] = 0;
      break;
    case "Minor":
      versionParts[1]++;
      versionParts[2] = 0;
      break;
    case "Patch":
      versionParts[2]++;
      break;
    default:
      throw new Error("Invalid version type");
  }

  return versionParts.join(".");
}
