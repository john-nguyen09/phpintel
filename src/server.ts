import { App } from "./app";
import { IConnection } from "vscode-languageserver";
import { InitializeProvider } from "./provider/initialize";
import { HoverProvider } from "./provider/hover";

App.run((connection: IConnection) => {
    connection.onInitialize(InitializeProvider.provide);
    connection.onHover(HoverProvider.provide);
});