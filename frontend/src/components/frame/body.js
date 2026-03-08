import React from 'react';
import {BodyHead} from "./body-head";
import {ReleasesScreen} from "../../window/ReleasesScreen";
import {ReleasesDetailScreen} from "../../window/ReleasesDetailScreen";
import {ReleaseCreateScreen} from "../../window/ReleaseCreateScreen";

const path_to_window = {
    'releases': {'window': ReleasesScreen, 'head': 'Релизы'},
    'release_card': {'window': ReleasesDetailScreen, 'head': ''},
    'release_create': {'window': ReleaseCreateScreen, 'head': 'Добавить/изменить релиз'},
}

export function Body({app}) {
    let simpleName = window.location.pathname.replaceAll('/','')
    let current_window = path_to_window[simpleName] || undefined
    if (current_window === undefined) {
        current_window = path_to_window['releases']
    }
    return (
        <div id={'body'} className={'body'}>
            <BodyHead text={current_window.head}/>
            {current_window.window(app)}
        </div>
    );
}