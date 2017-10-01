import { Injectable } from '@angular/core'
import { Router } from '@angular/router'
import { Observable } from "rxjs";
import { APIService, Serializable } from "./api";
import { TokenService } from "./auth";
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/switchMap'
import 'rxjs/add/operator/publishReplay'

export class Features extends Serializable {
	i18n:       boolean
	search:     boolean
	extractor:  boolean
	proxyHTTP:  boolean
	popularity: boolean
}

@Injectable()
export class FeaturesService {
    private features: Observable<Features>;

    constructor(
        private api: APIService,
        private tokenService: TokenService,
    ) {
        var features = this.tokenService.tokenObservable().switchMap(token =>
            this.api.get("features")
                .map(response =>
                    new Features().fromJSON(response.json().features)
                )
        ).publishReplay(1);
        features.connect();

        this.features = features;
    }

    getFeatures() : Observable<Features> {
        return this.features;
    }
}