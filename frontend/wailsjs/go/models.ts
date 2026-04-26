export namespace compiler {
	
	export class Info {
	    Found: boolean;
	    Path: string;
	    Version: string;
	
	    static createFrom(source: any = {}) {
	        return new Info(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Found = source["Found"];
	        this.Path = source["Path"];
	        this.Version = source["Version"];
	    }
	}

}

