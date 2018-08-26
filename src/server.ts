import { createConnection, TextDocumentSyncKind } from "vscode-languageserver";
import { HoverProvider } from "./provider/hover";
import { InitializeProvider } from "./provider/initialize";
import { Logger } from "./util/logger";
import { Hasher } from "./util/hash";

let connection = createConnection();

Logger.init(connection);
Hasher.init();

connection.onInitialize(InitializeProvider.provide);
connection.onHover(HoverProvider.provide);

connection.listen();