export namespace main {
	
	export class BlockedApp {
	    executableName: string;
	    displayName: string;
	
	    static createFrom(source: any = {}) {
	        return new BlockedApp(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.executableName = source["executableName"];
	        this.displayName = source["displayName"];
	    }
	}
	export class WindowInfo {
	    title: string;
	    exe: string;
	
	    static createFrom(source: any = {}) {
	        return new WindowInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.exe = source["exe"];
	    }
	}

}

