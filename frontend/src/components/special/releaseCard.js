import {make_cell} from "../trivial/cell";
import {too_many_requests} from "../trivial/error";

export function ReleaseCard({app}) {

    if ( app.state.oneData.error){
        return too_many_requests()
    }

    return <div >
                <div className={'rowFlex'}>
                    <div className={'new_card_top_row odd'} style={{'width':'20%'}}>
                        {make_cell(app, app.state.oneData, 'common', 'platform')}
                    </div>
                    <div className={'new_card_top_row'} style={{'width':'60%'}}>
                        {app.state.oneData.version}
                    </div>
                    <div className={'new_card_top_row odd'} style={{'width':'20%'}}>
                        {make_cell(app, app.state.oneData, 'date', 'created_at')}
                    </div>
                </div>
                <div className={'new_card_content_box'} style={{color: '#e0e0ff'}}>
                    {app.state.oneData.changes}
                </div>
                <div className={'new_card_content_bot'}>
                    {make_cell(app, app.state.oneData, 'release_download', 'Скачать')}
                </div>
        </div>
}
