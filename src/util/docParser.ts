import { TypeName } from "../type/name";
const _DocParser = require("doc-parser");

export enum DocNodeKind {
    Var = 'var',
    Param = 'param',
    Global = 'global',
    Return = 'return',
}

export enum DocTypeKind {
    Collection = 'collection',
    Type = 'type'
}

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
                optional: true
            }
        ],
        return: [
            {
                property: 'type',
                parser: 'type',
                optional: false,
            },
            {
                property: 'description',
                parser: 'text',
                optional: true
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

export type DocNode = VarDocNode | ParamDocNode | GlobalDocNode | ReturnDocNode;
export type DocTypeNode = DocType | DocTypeCollection;

export interface VarDocNode {
    kind: DocNodeKind.Var;
    type: DocTypeNode;
    variable: string;
    description: string;
}

export interface ParamDocNode {
    kind: DocNodeKind.Param;
    type: DocTypeNode;
    name: string;
}

export interface GlobalDocNode {
    kind: DocNodeKind.Global;
    type: DocTypeNode;
    variable: string;
    description: string;
}

export interface ReturnDocNode {
    kind: DocNodeKind.Return;
    type: DocTypeNode;
    description: string;
}

export interface DocType {
    kind: DocTypeKind.Type;
    fqn: boolean;
    name: string;
}

export interface DocTypeCollection {
    kind: DocTypeKind.Collection;
    value: DocType;
    index: any;
}

export function toTypeName(typeNode: DocTypeNode | null): TypeName | null {
    let docType: DocType | null = null;
    let name = '';

    if (typeNode === null) {
        return null;
    }

    if (typeNode.kind == DocTypeKind.Type) {
        docType = typeNode;
    } else if (typeNode.kind == DocTypeKind.Collection) {
        docType = typeNode.value;
    }

    if (docType == null) {
        return null;
    }

    if (docType.fqn) {
        name += '\\';
    }

    name += docType.name;

    return new TypeName(name);
}
