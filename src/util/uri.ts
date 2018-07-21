import URI from 'vscode-uri';

export function pathToUri(filePath: string): string {
    filePath = filePath.split('\\').join('/').trim();
    let parts = filePath.split('/');
    // Don't %-encode the colon after a Windows drive letter
    let first = parts.shift();
    if (first.substr(-1) !== ':') {
        first = encodeURIComponent(first);
    }
    parts = parts.map((part) => {
        return encodeURIComponent(part);
    });
    parts.unshift(first);
    filePath = parts.join('/');
    
    return 'file:///' + filePath;
}

export function uriToPath(uri: string)
{
    return URI.parse(uri).fsPath;
}