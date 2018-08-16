import { createConnection } from "vscode-languageserver";
import { HoverProvider } from "./provider/hover";
import { InitializeProvider } from "./provider/initialize";

let connection = createConnection();

connection.onInitialize(InitializeProvider.provide);
connection.onHover(HoverProvider.provide);