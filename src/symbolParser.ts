import { TreeNode, isToken, isPhrase } from "./util/parseTree";
import { Token, Phrase, PhraseType, TokenType } from "php7parser";
import { Symbol, TokenSymbol, isConsumer, isTransform, isCollection, isDocBlockConsumer } from "./symbol/symbol";
import { PhpDocument } from "./symbol/phpDocument";
import { Class } from "./symbol/class/class";
import { ClassHeader } from "./symbol/class/header";
import { ClassExtend } from "./symbol/class/extend";
import { ClassImplement } from "./symbol/class/implement";
import { ClassTraitUse } from "./symbol/class/traitUse";
import { QualifiedNameList } from "./symbol/list/qualifiedNameList";
import { QualifiedName } from "./symbol/name/qualifiedName";
import { NamespaceName } from "./symbol/name/namespaceName";
import { FunctionCall } from "./symbol/function/functionCall";
import { ArgumentExpressionList } from "./symbol/argumentExpressionList";
import { ConstantAccess } from "./symbol/constant/constantAccess";
import { AdditiveExpression } from "./symbol/type/additiveExpression";
import { Constant } from "./symbol/constant/constant";
import { Identifier } from "./symbol/identifier";
import { ClassConstant } from "./symbol/constant/classConstant";
import { Function } from "./symbol/function/function";
import { Return } from "./symbol/type/return";
import { FunctionHeader } from "./symbol/function/functionHeader";
import { Parameter } from "./symbol/variable/parameter";
import { TypeDeclaration } from "./symbol/type/typeDeclaration";
import { SimpleVariable } from "./symbol/variable/simpleVariable";
import { ClassTypeDesignator } from "./symbol/class/typeDesignator";
import { Method } from "./symbol/function/method";
import { MethodHeader } from "./symbol/function/methodHeader";
import { Property } from "./symbol/variable/property";
import { PropertyInitialiser } from "./symbol/variable/propertyInitialiser";
import { MemberModifierList } from "./symbol/class/memberModifierList";
import { PropertyDeclaration } from "./symbol/variable/propertyDeclaration";
import { DocBlock } from "./symbol/docBlock";

export class SymbolParser {
    protected symbolStack: Symbol[] = [];
    protected doc: PhpDocument;
    protected lastDocBlock: DocBlock = null

    constructor(doc: PhpDocument) {
        this.doc = doc;
        this.pushSymbol(this.doc);
    }

    traverse(tree: Phrase) {
        let depth = 0;

        this.realTraverse(tree, depth);
    }

    private realTraverse(node: TreeNode, depth: number) {
        this.preorder(node, depth);

        if ('children' in node) {
            for (let child of node.children) {
                this.realTraverse(child, depth + 1);
            }
        }

        this.postorder(node, depth);
    }

    public getTree(): Symbol {
        return this.doc;
    }

    getParentSymbol(): Symbol {
        return this.symbolStack[this.symbolStack.length - 1];
    }

    pushSymbol(symbol: Symbol) {
        this.symbolStack.push(symbol);
    }

    preorder(node: TreeNode, depth: number) {
        let parentSymbol = this.getParentSymbol();

        if (isToken(node)) {
            if (node.tokenType == TokenType.DocumentComment) {
                this.lastDocBlock = new DocBlock(node, this.doc, depth);
            } else {
                let symbol = new TokenSymbol(node, this.doc);

                if (parentSymbol && isConsumer(parentSymbol)) {
                    parentSymbol.consume(symbol);
                }
            }
        } else if (isPhrase(node)) {
            let p = <Phrase>node;

            switch (p.phraseType) {
                case PhraseType.NamespaceName:
                    this.pushSymbol(new NamespaceName(node, this.doc));
                    break;
                case PhraseType.QualifiedName:
                    this.pushSymbol(new QualifiedName(node, this.doc));
                    break;
                case PhraseType.QualifiedNameList:
                    this.pushSymbol(new QualifiedNameList(p, this.doc));
                    break;
                case PhraseType.Identifier:
                    this.pushSymbol(new Identifier(p, this.doc));
                    break;

                case PhraseType.ClassDeclaration:
                    this.pushSymbol(new Class(p, this.doc));
                    break;
                case PhraseType.ClassDeclarationHeader:
                    this.pushSymbol(new ClassHeader(p, this.doc));
                    break;
                case PhraseType.ClassBaseClause:
                    this.pushSymbol(new ClassExtend(p, this.doc));
                    break;
                case PhraseType.ClassInterfaceClause:
                    this.pushSymbol(new ClassImplement(p, this.doc));
                    break;
                case PhraseType.TraitUseClause:
                    this.pushSymbol(new ClassTraitUse(p, this.doc));
                    break;

                case PhraseType.ConstElement:
                    this.pushSymbol(new Constant(p, this.doc));
                    break;
                case PhraseType.FunctionCallExpression:
                    this.pushSymbol(new FunctionCall(p, this.doc));
                    break;
                case PhraseType.ClassConstElement:
                    this.pushSymbol(new ClassConstant(p, this.doc));
                    break;
                case PhraseType.ArgumentExpressionList:
                    this.pushSymbol(new ArgumentExpressionList(p, this.doc));
                    break;
                case PhraseType.ConstantAccessExpression:
                    this.pushSymbol(new ConstantAccess(p, this.doc));
                    break;
                case PhraseType.AdditiveExpression:
                    this.pushSymbol(new AdditiveExpression(p, this.doc));
                    break;

                case PhraseType.FunctionDeclaration:
                    this.pushSymbol(new Function(p, this.doc));
                    break;
                case PhraseType.FunctionDeclarationHeader:
                    this.pushSymbol(new FunctionHeader(p, this.doc));
                    break;
                case PhraseType.MethodDeclaration:
                    this.pushSymbol(new Method(p, this.doc));
                    break;
                case PhraseType.MethodDeclarationHeader:
                    this.pushSymbol(new MethodHeader(p, this.doc));
                    break;
                case PhraseType.ParameterDeclaration:
                    this.pushSymbol(new Parameter(p, this.doc));
                    break;
                case PhraseType.TypeDeclaration:
                    this.pushSymbol(new TypeDeclaration(p, this.doc));
                    break;
                case PhraseType.ReturnStatement:
                    this.pushSymbol(new Return(p, this.doc));
                    break;
                case PhraseType.SimpleVariable:
                    this.pushSymbol(new SimpleVariable(p, this.doc));
                    break;
                case PhraseType.ClassTypeDesignator:
                    this.pushSymbol(new ClassTypeDesignator(p, this.doc));

                case PhraseType.PropertyElement:
                    this.pushSymbol(new Property(p, this.doc));
                    break;
                case PhraseType.PropertyInitialiser:
                    this.pushSymbol(new PropertyInitialiser(p, this.doc));
                    break;
                case PhraseType.MemberModifierList:
                    this.pushSymbol(new MemberModifierList(p, this.doc));
                    break;
                case PhraseType.PropertyDeclaration:
                    this.pushSymbol(new PropertyDeclaration(p, this.doc));
                    break;

                default:
                    this.pushSymbol(null);
            }
        }
    }

    postorder(node: TreeNode, depth: number) {
        if (isToken(node)) {
            return;
        }

        let symbol = this.symbolStack.pop();

        if (isTransform(symbol) && symbol.realSymbol) {
            symbol = symbol.realSymbol;
        }

        if (symbol == null) {
            return;
        }

        for (let i = this.symbolStack.length - 1; i >= 0; i--) {
            let prev = this.symbolStack[i];

            if (!prev || !isConsumer(prev)) {
                continue;
            }

            if (isCollection(symbol)) {
                let isConsumed = false;

                for (let realSymbol of symbol.realSymbols) {
                    if (!realSymbol) {
                        continue;
                    }

                    isConsumed = prev.consume(realSymbol) || isConsumed;
                }

                if (isConsumed) {
                    break;
                }
            } else {
                if (prev.consume(symbol)) {
                    break;
                }
            }
        }

        if (isDocBlockConsumer(symbol) && this.lastDocBlock != null) {
            symbol.consumeDocBlock(this.lastDocBlock);

            this.lastDocBlock = null;
        }
    }
}
