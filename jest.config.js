/** @type {import('jest').Config} */
module.exports = {
  rootDir: ".",
  moduleDirectories: ["node_modules", "frontend"],
  testEnvironment: "jest-fixed-jsdom",
  moduleNameMapper: {
    "\\.(css|scss|sass)$": "identity-obj-proxy",
  },
  testMatch: ["<rootDir>/frontend/**/*tests.[jt]s?(x)"],
  setupFilesAfterEnv: ["<rootDir>/frontend/test/test-setup.ts"],
  testEnvironmentOptions: {
    url: "http://localhost/",
    customExportConditions: [""],
  },
  transform: {
    "^.+\\.tsx?$": ["ts-jest", { diagnostics: false }],
    "^.+\\.m?js$": "babel-jest",
  },
  transformIgnorePatterns: [
    "/node_modules/(?!msw|@mswjs|@bundled-es-modules|@open-draft|headers-polyfill|outvariant|strict-event-emitter|rettime|undici)/",
  ],
};
