import { Symbol, Consumer, DocBlockConsumer } from "../symbol";
import { Expression } from "../type/expression";
import { TypeComposite } from "../../type/composite";
import { DocBlock } from "../docBlock";
import { DocNodeKind, toTypeName, VarDocNode } from "../../util/docParser";
import { Location } from "../meta/location";
import { Reference, RefKind } from "../reference";
import { TypeName } from "../../type/name";

export class Variable extends Symbol implements Consumer, DocBlockConsumer, Reference {
    public readonly refKind = RefKind.Variable;
    public type: TypeComposite = new TypeComposite();
    public location: Location = new Location();
    public scope: TypeName | null = null;

    protected expression: Expression;

    constructor(public name: string, type?: TypeComposite) {
        super();

        if (type) {
            this.type = type;
        }

        this.expression = new Expression();
    }

    consume(other: Symbol) {
        let result = this.expression.consume(other);

        if (!this.expression.type.isEmptyName()) {
            this.type.push(this.expression.type);
        }

        return result;
    }

    consumeDocBlock(doc: DocBlock) {
        let docAst = doc.docAst;
        if (docAst.kind == 'doc') {
            let varDocNodes = doc.getNodes<VarDocNode>(DocNodeKind.Var);

            for (let i = 0; i < varDocNodes.length; i++) {
                let isThisVar = false;

                if (varDocNodes[i].variable == null) {
                    isThisVar = true;
                } else {
                    let docVarName = '$' + varDocNodes[i].variable;

                    if (this.name == docVarName) {
                        isThisVar = true;
                    }
                }

                if (isThisVar) {
                    let typeName = toTypeName(varDocNodes[i].type);

                    if (typeName != null) {
                        this.type.push(typeName);
                    }
                }
            }
        }
    }

    get refName(): string {
        return this.name;
    }
}