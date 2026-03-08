import {get_releases_total, makePreparations} from "../store/preparations";
import {releasesPlatform} from "../store/filters";
import {releases} from "../store/tabs";
import {Table} from "../components/trivial/table";

export function ReleasesScreen(app) {

    if (makePreparations(app, [get_releases_total])){
        return
    }

    const filters = [
        releasesPlatform()
    ]
    const tabs = [
        releases,
    ]

    if (app.state.isStuff) {
        return <div>
            <button onClick={()=>{window.location='release_create'}}>+ Добавить</button>
            <Table app={app} data={app.state.data} tabs={tabs} filters={filters} large={true}/>
        </div>
    } else{
        return <Table app={app} data={app.state.data} tabs={tabs} filters={filters} large={true}/>
    }


}