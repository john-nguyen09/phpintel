import { TokenSymbol } from "./symbol";
import { PhpDocument } from "./phpDocument";
import { Token } from "php7parser";
import { DocParser, DocAst, DocNode } from "../docParser";

export class DocBlock extends TokenSymbol {
    public docAst: DocAst;

    constructor(token: Token, doc: PhpDocument, public depth: number) {
        super(token, doc);

        this.docAst = DocParser.parse(this.text);
    }

    public getNodes<T extends DocNode>(kind: string): T[] {
        let nodes: T[] = [];

        for (let node of this.docAst.body) {
            if (node.kind == kind) {
                nodes.push(<T>node);
            }
        }

        return nodes;
    }

    public static isType<T extends DocNode>(docNode: DocNode, type: string): docNode is T {
        return docNode.kind == type;
    }
}
