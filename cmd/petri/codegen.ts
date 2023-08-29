import type { CodegenConfig } from "@graphql-codegen/cli";


const config: CodegenConfig = {
  schema: "../petrid/graph/**/*.graphqls",
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