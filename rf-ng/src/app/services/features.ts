import { Injectable } from '@angular/core'
import { Router } from '@angular/router'
import { Observable, ConnectableObservable } from "rxjs";
import { APIService } from "./api";
import { TokenService } from "./auth";
import { switchMap, publishReplay, map } from 'rxjs/operators';



interface featuresPayload {
    features: Features
}

export interface Features {
	i18n:       boolean;
	search:     boolean;
	extractor:  boolean;
	proxyHTTP:  boolean;
	popularity: boolean;
}

@Injectable({providedIn: "root"})
export class FeaturesService {
    private features: Observable<Features>;

    constructor(
        private api: APIService,
        private tokenService: TokenService,
    ) {
        var features = this.tokenService.tokenObservable().pipe(
            switchMap(token => this.api.get<featuresPayload>("features").pipe(
                map(p => p.features)
            )),
            publishReplay<Features>(1)
        ) as ConnectableObservable<Features>;
        features.connect();

        this.features = features;
    }

    getFeatures() : Observable<Features> {
        return this.features;
    }
}