export namespace good {
	
	export class RecipeMaterial {
	    name: string;
	    cnt: number;
	
	    static createFrom(source: any = {}) {
	        return new RecipeMaterial(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.cnt = source["cnt"];
	    }
	}
	export class RecipeDrop {
	    cnt: number;
	    prob: number;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new RecipeDrop(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cnt = source["cnt"];
	        this.prob = source["prob"];
	        this.name = source["name"];
	    }
	}
	export class ItemRecipe {
	    drop: RecipeDrop[];
	    materials: RecipeMaterial[];
	    focus: number;
	
	    static createFrom(source: any = {}) {
	        return new ItemRecipe(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.drop = this.convertValues(source["drop"], RecipeDrop);
	        this.materials = this.convertValues(source["materials"], RecipeMaterial);
	        this.focus = source["focus"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Good {
	    sidebar?: number;
	    tab?: number;
	    name: string;
	    range: number[];
	    price: number;
	    category1?: string;
	    category2?: string;
	    private: boolean;
	    recipe: ItemRecipe[];
	
	    static createFrom(source: any = {}) {
	        return new Good(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sidebar = source["sidebar"];
	        this.tab = source["tab"];
	        this.name = source["name"];
	        this.range = source["range"];
	        this.price = source["price"];
	        this.category1 = source["category1"];
	        this.category2 = source["category2"];
	        this.private = source["private"];
	        this.recipe = this.convertValues(source["recipe"], ItemRecipe);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Profit {
	    name: string;
	    cnt: number;
	    cost: number;
	    price: number;
	    byProducts: Profit[];
	    value: number;
	    comment: string;
	    num: number;
	
	    static createFrom(source: any = {}) {
	        return new Profit(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.cnt = source["cnt"];
	        this.cost = source["cost"];
	        this.price = source["price"];
	        this.byProducts = this.convertValues(source["byProducts"], Profit);
	        this.value = source["value"];
	        this.comment = source["comment"];
	        this.num = source["num"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

