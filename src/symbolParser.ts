import { TreeNode, isToken, isPhrase } from "./util/parseTree";
import { Token, Phrase, TokenType, PhraseType } from "php7parser";
import { Symbol, TokenSymbol, TransformSymbol } from "./symbol/symbol";
import { PhpDocument } from "./phpDocument";
import { Class } from "./symbol/class/class";
import { ClassHeader } from "./symbol/class/header";
import { ClassExtend } from "./symbol/class/extend";
import { ClassImplement } from "./symbol/class/implement";
import { ClassTraitUse } from "./symbol/class/traitUse";
import { QualifiedNameList } from "./symbol/list/qualifiedNameList";
import { File } from "./symbol/file";
import { QualifiedName } from "./symbol/name/qualifiedName";
import { NamespaceName } from "./symbol/name/namespaceName";
import { FunctionCall } from "./symbol/functionCall";
import { ArgumentExpressionList } from "./symbol/argumentExpressionList";
import { ConstantAccess } from "./symbol/constant/constantAccess";
import { AdditiveExpression } from "./symbol/type/additiveExpression";

export class SymbolParser {
    protected symbolStack: Symbol[] = [];
    protected doc: PhpDocument;
    protected file: Symbol;
    protected spine: TreeNode[];

    constructor(doc: PhpDocument) {
        this.doc = doc;

        this.file = new File();
        this.symbolStack.push(this.file);
    }

    traverse(tree: Phrase) {
        this.spine = [];
        this.realTraverse(tree, this.spine);
    }

    private realTraverse(node: TreeNode, spine: TreeNode[]) {
        this.preorder(node, spine);

        if ('children' in node) {
            spine.push(node);
            for (let child of node.children) {
                this.realTraverse(child, spine);
            }
            spine.pop();
        }

        this.postorder(node, spine);
    }

    public getTree(): Symbol {
        return this.file;
    }

    getParentSymbol(): Symbol {
        return this.symbolStack[this.symbolStack.length - 1];
    }

    pushSymbol(symbol: Symbol) {
        this.symbolStack.push(symbol);
    }

    preorder(node: TreeNode, spine: TreeNode[]) {
        let parentSymbol = this.getParentSymbol();

        if (isToken(node)) {
            let t = <Token>node;

            if (parentSymbol) {
                parentSymbol.consume(new TokenSymbol(t, this.doc));
            }
        } else if (isPhrase(node)) {
            let p = <Phrase>node;

            switch(p.phraseType) {
                case PhraseType.NamespaceName:
                    this.pushSymbol(new NamespaceName(node));
                    break;
                case PhraseType.QualifiedName:
                    this.pushSymbol(new QualifiedName(node));
                    break;
                case PhraseType.QualifiedNameList:
                    this.pushSymbol(new QualifiedNameList(p));
                    break;

                case PhraseType.ClassDeclaration:
                    this.pushSymbol(new Class(p, this.doc));
                    break;
                case PhraseType.ClassDeclarationHeader:
                    this.pushSymbol(new ClassHeader(p));
                    break;
                case PhraseType.ClassBaseClause:
                    this.pushSymbol(new ClassExtend(p));
                    break;
                case PhraseType.ClassInterfaceClause:
                    this.pushSymbol(new ClassImplement(p));
                    break;
                case PhraseType.TraitUseClause:
                    this.pushSymbol(new ClassTraitUse(p));
                    break;

                case PhraseType.FunctionCallExpression:
                    this.pushSymbol(new FunctionCall(p));
                    break;
                case PhraseType.ArgumentExpressionList:
                    this.pushSymbol(new ArgumentExpressionList(p));
                    break;
                case PhraseType.ConstantAccessExpression:
                    this.pushSymbol(new ConstantAccess(p));
                    break;
                case PhraseType.AdditiveExpression:
                    this.pushSymbol(new AdditiveExpression(p));
                    break;

                default:
                    this.pushSymbol(null);
            }
        }
    }

    postorder(node: TreeNode, spine: TreeNode[]) {
        if (isToken(node)) {
            return;
        }

        let symbol = this.symbolStack.pop();

        if (symbol !== null && 'realSymbol' in symbol) {
            symbol = (<TransformSymbol>symbol).realSymbol;
        }

        if (symbol === null) {
            return;
        }

        for (let i = this.symbolStack.length - 1; i >= 0; i--) {
            if (this.symbolStack[i] && this.symbolStack[i].consume(symbol)) {
                break;
            }
        }
    }
}