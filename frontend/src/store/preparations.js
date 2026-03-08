import {getAllNews} from "../api/newsAPI";
import {getDetailNews} from "../api/newsAPI";
import {getAllReleases, getDetailRelease} from "../api/releasesAPI";
import {getActiveGames, getGeneralStat} from "../api/statAPI";

export const get_news_total = 'get_news_total'
export const get_news_detail = 'get_news_detail'
export const get_releases_total = 'get_releases_total'
export const get_releases_detail = 'get_releases_detail'
export const get_stat_total = 'get_stat_total'
export const get_games_active = 'get_games_active'


function getURLParams() {return new URLSearchParams(window.location.search)}

function getID(noDefault=false){
    const urlParams = getURLParams();
    return noDefault ? urlParams.get('id') : urlParams.get('id') || 1
}

const no_many_data = (app) => {return app.state.data === null}
const no_one_data = (app) => {return app.state.oneData === undefined}


const query_all_news = (app) => {getAllNews(app, app.state.page || 1, app.state.sortIndex, app.state.sortOrder, app.state.filters)}
const query_detail_news = (app) => {getDetailNews(app, getID(true))}

const query_all_releases = (app) => {getAllReleases(app, app.state.page || 1, app.state.sortIndex, app.state.sortOrder, app.state.filters)}
const query_detail_release = (app) => {getDetailRelease(app, getID(true))}

const query_stat_general = (app) => {getGeneralStat(app)}
const query_games_active = (app) => {getActiveGames(app)}

const functionsForPrepare = {
    get_news_total: {'test': no_many_data, 'get': query_all_news},
    get_news_detail: {'test': no_one_data, 'get': query_detail_news},
    get_releases_total: {'test': no_many_data, 'get': query_all_releases},
    get_releases_detail: {'test': no_one_data, 'get': query_detail_release},
    get_stat_total: {'test': no_one_data, 'get': query_stat_general},
    get_games_active: {'test': no_many_data, 'get': query_games_active},

}

export function makePreparations(app, dataNeeded){
    for (let i = 0; i < dataNeeded.length; i++) {
        if(functionsForPrepare[dataNeeded[i]]['test'](app)){
            functionsForPrepare[dataNeeded[i]]['get'](app)
            return true
        }
    }
    return false
}
