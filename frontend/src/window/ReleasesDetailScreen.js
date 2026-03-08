import {get_releases_detail, makePreparations} from "../store/preparations";
import {ReleaseCard} from "../components/special/releaseCard";
import {deleteNews} from "../api/newsAPI";
import {CalendarBase} from "../components/special/newCard";
import {deleteRelease} from "../api/releasesAPI";

export function ReleasesDetailScreen(app) {

    if (makePreparations(app, [get_releases_detail])){
        return
    }

    if (app.state.isStuff) {
        return <div>
            <button onClick={()=>{
                deleteRelease(app, app.state.oneData.id)
                // window.location='/releases'
            }}>- Удалить</button>
            <button onClick={()=>{
                window.location.href = ('/release_create?id=' + app.state.oneData.id);
            }}>* Изменить</button>
            <ReleaseCard  app={app}/>
        </div>
    } else{
        return <ReleaseCard  app={app}/>
    }
}