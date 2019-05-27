import { Symbol, Consumer, NamedSymbol, Locatable, DocBlockConsumer } from "../symbol";
import { TypeName } from "../../type/name";
import { Location } from "../meta/location";
import { InterfaceHeader } from "./header";
import { DocBlock } from "../docBlock";
import { ImportTable } from "../../type/importTable";

export class Interface extends Symbol implements Consumer, DocBlockConsumer, NamedSymbol, Locatable {
    public name: TypeName;
    public description: string;
    public parents: TypeName[] = [];
    public location: Location;

    public consume(other: Symbol) {
        if (other instanceof InterfaceHeader) {
            this.name = other.name;
        }

        return false;
    }

    public consumeDocBlock(docBlock: DocBlock): void {
        this.description = docBlock.docAst.summary;
    }

    public resolveName(importTable: ImportTable): void {
        this.name.resolveDefinitionToFqn(importTable);
    }

    public getName(): string {
        return this.name.toString();
    }
}