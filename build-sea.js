const { copyFileSync, readFileSync } = require("fs");
const { execSync } = require("child_process");

const platform = process.platform;
const outputName = platform === "win32" ? "loops.exe" : "loops";
const blob = "dist-sea/loops.blob";

console.log("building single executable application...");

try {
  console.log("1. copying node.js binary...");
  copyFileSync(process.execPath, outputName);

  console.log("2. injecting application blob...");
  if (platform === "darwin") {
    execSync(`codesign --remove-signature ${outputName}`);
    execSync(
      `npx postject ${outputName} NODE_SEA_BLOB ${blob} --sentinel-fuse NODE_SEA_FUSE_fce680ab2cc467b6e072b8b5df1996b2 --macho-segment-name NODE_SEA`,
    );
    console.log("3. signing binary...");
    execSync(`codesign --sign - ${outputName}`);
  } else if (platform === "linux") {
    execSync(
      `npx postject ${outputName} NODE_SEA_BLOB ${blob} --sentinel-fuse NODE_SEA_FUSE_fce680ab2cc467b6e072b8b5df1996b2`,
    );
  } else if (platform === "win32") {
    execSync(
      `npx postject ${outputName} NODE_SEA_BLOB ${blob} --sentinel-fuse NODE_SEA_FUSE_fce680ab2cc467b6e072b8b5df1996b2`,
    );
  }

  console.log(`✓ single executable created: ${outputName}`);
  console.log(`  test it with: ./${outputName} --help`);
} catch (error) {
  console.error("error building sea:", error.message);
  process.exit(1);
}
