import type { CodegenConfig } from "@graphql-codegen/cli";

const config: CodegenConfig = {
  schema: "http://localhost:8081/graphql",
  documents: ["../petrid/graph/**/*.graphql"],
  generates: {
    "./app/models/__generated__/": {
      preset: "client",
      plugins: [],
      presetConfig: {
        gqlTagName: "gql"
      }
    }
  },
  ignoreNoDocuments: true
};


export default config;