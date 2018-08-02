const _DocParser = require("doc-parser");

export namespace DocParser {
    const customDocBlocks = {
        global: [
            {
                property: 'type',
                parser: 'type',
                optional: true
            },
            {
                property: 'variable',
                parser: 'variable',
                optional: true
            },
            {
                property: 'description',
                parser: 'text',
                optional: true,
                default: ''
            }
        ],
        var: [
            {
                property: 'type',
                parser: 'type',
                optional: false
            },
            {
                property: 'variable',
                parser: 'variable',
                optional: true
            },
            {
                property: 'description',
                parser: 'text',
                optional: true,
                default: ''
            }
        ]
    };
    const docParser = new _DocParser(customDocBlocks);

    export function parse(comment: string): DocAst {
        return docParser.parse(comment);
    }
}

export interface DocAst {
    kind: string;
    summary: string,
    body: DocNode[]
}

export interface DocNode {
    kind: string;
}

export interface VarDocNode extends DocNode {
    type: DocType;
    variable: string;
    description: string;
}

export interface DocType {
    kind: string;
    fqn: boolean;
    name: string;
}