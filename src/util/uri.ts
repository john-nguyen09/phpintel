import URI from 'vscode-uri/lib/umd';
import * as path from 'path';

export function pathToUri(filePath: string): string {
    filePath = path.resolve(filePath).replace(/\\/g, '/');

    if (filePath[0] != '/') {
        filePath = '/' + filePath;
    }
    
    return encodeURI('file://' + filePath);
}

export function uriToPath(uri: string) {
    return URI.parse(uri).fsPath;
}

export function toRelative(uri: string) {
    let baseUri = pathToUri(path.resolve(__dirname, '..' , '..'));

    if (uri.indexOf(baseUri) === 0) {
        return uri.substr(baseUri.length + 1);
    }

    return uri;
}