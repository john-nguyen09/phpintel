import { TreeNode, isToken, isPhrase, nodeRange } from "../util/parseTree";
import { Phrase } from "php7parser";
import { Symbol, TokenSymbol, isConsumer, isTransform, isCollection, isDocBlockConsumer, isLocatable } from "./symbol";
import { PhpDocument } from "./phpDocument";
import { Class } from "./class/class";
import { ClassHeader } from "./class/header";
import { ClassExtend } from "./class/extend";
import { ClassImplement } from "./class/implement";
import { ClassTraitUse } from "./class/traitUse";
import { QualifiedNameList } from "./list/qualifiedNameList";
import { QualifiedName } from "./name/qualifiedName";
import { NamespaceName } from "./name/namespaceName";
import { FunctionCall } from "./function/functionCall";
import { ArgumentExpressionList } from "./argumentExpressionList";
import { ConstantAccess } from "./constant/constantAccess";
import { AdditiveExpression } from "./type/additiveExpression";
import { Constant } from "./constant/constant";
import { Identifier } from "./identifier";
import { ClassConstant } from "./constant/classConstant";
import { Function } from "./function/function";
import { Return } from "./type/return";
import { FunctionHeader } from "./function/functionHeader";
import { Parameter } from "./variable/parameter";
import { TypeDeclaration } from "./type/typeDeclaration";
import { SimpleVariable } from "./variable/simpleVariable";
import { ClassTypeDesignator } from "./class/typeDesignator";
import { Method } from "./function/method";
import { MethodHeader } from "./function/methodHeader";
import { Property } from "./variable/property";
import { PropertyInitialiser } from "./variable/propertyInitialiser";
import { MemberModifierList } from "./class/memberModifierList";
import { PropertyDeclaration } from "./variable/propertyDeclaration";
import { DocBlock } from "./docBlock";
import { NamespaceDefinition } from "./namespace/definition";
import { NamespaceUse } from "./namespace/Use";
import { NamespaceUseClause } from "./namespace/useClause";
import { NamespaceAliasClause } from "./namespace/aliasClause";
import { PhraseKind, TokenKind } from "../util/parser";
import { VariableAssignment } from "./variable/varibleAssignment";
import { Visitor } from "../treeTraverser";
import { Location } from "./meta/location";

export class SymbolParser implements Visitor<TreeNode> {
    protected symbolStack: (Symbol | null)[] = [];
    protected doc: PhpDocument;
    protected lastDocBlock: DocBlock | null = null;

    constructor(doc: PhpDocument) {
        this.doc = doc;
        this.pushSymbol(this.doc);
    }

    public getTree(): PhpDocument {
        return this.doc;
    }

    getParentSymbol(): Symbol | null {
        return this.symbolStack[this.symbolStack.length - 1];
    }

    pushSymbol(symbol: Symbol | null) {
        this.symbolStack.push(symbol);
        this.doc.pushSymbol(symbol);
    }

    preorder(node: TreeNode) {
        let parentSymbol = this.getParentSymbol();

        if (isToken(node)) {
            let tokenType: number = <number>node.tokenType;

            if (tokenType == TokenKind.DocumentComment) {
                this.lastDocBlock = new DocBlock(node, this.doc);
            } else {
                let symbol = new TokenSymbol(node, this.doc);

                if (parentSymbol && isConsumer(parentSymbol)) {
                    parentSymbol.consume(symbol);
                }
            }
        } else if (isPhrase(node)) {
            const p = <Phrase>node;
            const phraseType: number = <number>p.phraseType;

            switch (phraseType) {
                case PhraseKind.NamespaceDefinition:
                    this.pushSymbol(new NamespaceDefinition());
                    break;
                case PhraseKind.NamespaceName:
                    this.pushSymbol(new NamespaceName());
                    break;
                case PhraseKind.QualifiedName:
                    this.pushSymbol(new QualifiedName());
                    break;
                case PhraseKind.QualifiedNameList:
                    this.pushSymbol(new QualifiedNameList());
                    break;
                case PhraseKind.Identifier:
                    this.pushSymbol(new Identifier());
                    break;

                case PhraseKind.NamespaceUseDeclaration:
                    this.pushSymbol(new NamespaceUse());
                    break;
                case PhraseKind.NamespaceUseClause:
                    this.pushSymbol(new NamespaceUseClause());
                    break;
                case PhraseKind.NamespaceUseGroupClause:
                    this.pushSymbol(new NamespaceUseClause());
                    break;
                case PhraseKind.NamespaceAliasingClause:
                    this.pushSymbol(new NamespaceAliasClause());
                    break;

                case PhraseKind.ClassDeclaration:
                    this.pushSymbol(new Class());
                    break;
                case PhraseKind.ClassDeclarationHeader:
                    this.pushSymbol(new ClassHeader());
                    break;
                case PhraseKind.ClassBaseClause:
                    this.pushSymbol(new ClassExtend());
                    break;
                case PhraseKind.ClassInterfaceClause:
                    this.pushSymbol(new ClassImplement());
                    break;
                case PhraseKind.TraitUseClause:
                    this.pushSymbol(new ClassTraitUse());
                    break;

                case PhraseKind.ConstElement:
                    this.pushSymbol(new Constant());
                    break;
                case PhraseKind.FunctionCallExpression:
                    this.pushSymbol(new FunctionCall());
                    break;
                case PhraseKind.ClassConstElement:
                    this.pushSymbol(new ClassConstant());
                    break;
                case PhraseKind.ArgumentExpressionList:
                    this.pushSymbol(new ArgumentExpressionList());
                    break;
                case PhraseKind.ConstantAccessExpression:
                    this.pushSymbol(new ConstantAccess());
                    break;
                case PhraseKind.AdditiveExpression:
                    this.pushSymbol(new AdditiveExpression());
                    break;

                case PhraseKind.FunctionDeclaration:
                    this.pushSymbol(new Function());
                    break;
                case PhraseKind.FunctionDeclarationHeader:
                    this.pushSymbol(new FunctionHeader());
                    break;
                case PhraseKind.MethodDeclaration:
                    this.pushSymbol(new Method());
                    break;
                case PhraseKind.MethodDeclarationHeader:
                    this.pushSymbol(new MethodHeader());
                    break;
                case PhraseKind.ParameterDeclaration:
                    this.pushSymbol(new Parameter());
                    break;
                case PhraseKind.TypeDeclaration:
                    this.pushSymbol(new TypeDeclaration());
                    break;
                case PhraseKind.ReturnStatement:
                    this.pushSymbol(new Return());
                    break;
                case PhraseKind.SimpleAssignmentExpression:
                    this.pushSymbol(new VariableAssignment());
                    break;
                case PhraseKind.SimpleVariable:
                    this.pushSymbol(new SimpleVariable());
                    break;
                case PhraseKind.ClassTypeDesignator:
                    this.pushSymbol(new ClassTypeDesignator());
                    break;
                case PhraseKind.PropertyElement:
                    this.pushSymbol(new Property());
                    break;
                case PhraseKind.PropertyInitialiser:
                    this.pushSymbol(new PropertyInitialiser());
                    break;
                case PhraseKind.MemberModifierList:
                    this.pushSymbol(new MemberModifierList());
                    break;
                case PhraseKind.PropertyDeclaration:
                    this.pushSymbol(new PropertyDeclaration());
                    break;

                default:
                    this.pushSymbol(null);
            }

            let symbol = this.symbolStack[this.symbolStack.length - 1];

            if (symbol !== null) {
                if (isDocBlockConsumer(symbol) && this.lastDocBlock != null) {
                    symbol.consumeDocBlock(this.lastDocBlock);
    
                    this.lastDocBlock = null;
                }
                
                if (isLocatable(symbol)) {
                    symbol.location = new Location(this.doc.uri, nodeRange(node, this.doc.text));
                }
            }
        }
    }

    postorder(node: TreeNode) {
        if (isToken(node)) {
            return;
        }

        let symbol = this.symbolStack.pop();

        if (symbol && isTransform(symbol) && symbol.realSymbol) {
            symbol = symbol.realSymbol;
        }

        if (symbol == null || symbol == undefined) {
            return;
        }

        let isConsumed: boolean = false;
        for (let i = this.symbolStack.length - 1; i >= 0; i--) {
            let prev = this.symbolStack[i];

            if (!prev || !isConsumer(prev)) {
                continue;
            }

            if (isCollection(symbol)) {
                for (let realSymbol of symbol.realSymbols) {
                    if (!realSymbol) {
                        continue;
                    }

                    isConsumed = prev.consume(realSymbol) || isConsumed;
                }
            } else {
                isConsumed = prev.consume(symbol);
            }

            if (isConsumed) {
                break;
            }
        }
    }
}
