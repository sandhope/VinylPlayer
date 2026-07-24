export namespace main {
	
	export class Track {
	    id: string;
	    path: string;
	    title: string;
	    artist: string;
	    album: string;
	    format: string;
	    url: string;
	    coverUrl: string;
	    lyricUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new Track(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.path = source["path"];
	        this.title = source["title"];
	        this.artist = source["artist"];
	        this.album = source["album"];
	        this.format = source["format"];
	        this.url = source["url"];
	        this.coverUrl = source["coverUrl"];
	        this.lyricUrl = source["lyricUrl"];
	    }
	}

}

