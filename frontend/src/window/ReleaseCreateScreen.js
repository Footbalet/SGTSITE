import {get_releases_detail, makePreparations} from "../store/preparations";
import {backendURL} from "../api/API";

export function ReleaseCreateScreen(app) {

    if (makePreparations(app, [get_releases_detail])){
        return
    }

    return <CalendarBase app={app}/>
}

function CalendarBase({app}) {
    var data = app.state.oneData
    var error = data.error
    return <div >
        <div className={'rowFlex'}>
            <div className={'new_card_top_row odd'} style={{'width':'20%'}}>
                Платформа
            </div>
            <div className={'new_card_top_row'} style={{'width':'80%'}}>
                <input id={'platform'} defaultValue={error ?  'windows' : data.platform}/>
            </div>
        </div>
        <div className={'rowFlex'}>
            <div className={'new_card_top_row odd'} style={{'width':'20%'}}>
                Версия
            </div>
            <div className={'new_card_top_row'} style={{'width':'80%'}}>
                <input id ={'version'} defaultValue={error ? '' : data.version}/>
            </div>
        </div>
        <div className={'rowFlex'}>
            <div className={'new_card_top_row odd'} style={{'width': '20%'}}>
                Изменения
            </div>
            <div className={'new_card_top_row'} style={{'width': '80%'}}>
                <textarea id={'changes'} defaultValue={error ? '' : data.changes}/>
            </div>
        </div>
        <div className={'rowFlex'}>
            <div className={'new_card_top_row odd'} style={{'width':'20%'}}>
                Файл
            </div>
            <div className={'new_card_top_row'} style={{'width':'80%'}}>
                <input type="file" id="file"/>
            </div>
        </div>
        <button onClick={() => {
            if (error) {
                var form_data = new FormData()
                form_data.append('platform', document.getElementById('platform').value)
                form_data.append('version', document.getElementById('version').value)
                form_data.append('changes', document.getElementById('changes').value)
                const fileInput = document.getElementById('file');
                if (fileInput.files.length > 0) {
                    form_data.append('file', fileInput.files[0]);
                }
                fetch(
                    backendURL + '/releases/', {
                        method:'POST',
                        headers: {
                            'Authorization': localStorage.getItem('key_to_my_heart')
                        },
                        body: form_data
                    }
                )
            } else {
                var form_data = new FormData()
                form_data.append('platform', document.getElementById('platform').value)
                form_data.append('version', document.getElementById('version').value)
                form_data.append('changes', document.getElementById('changes').value)
                const fileInput = document.getElementById('file');
                if (fileInput.files.length > 0) {
                    form_data.append('file', fileInput.files[0]);
                }
                fetch(
                    backendURL + '/releases/'+app.state.oneData.id, {
                        method:'PUT',
                        headers: {
                            'Authorization': localStorage.getItem('key_to_my_heart')
                        },
                        body: form_data
                    }
                )
            }
        }}>{error ? 'Добавить' : 'Сохранить'}</button>
    </div>
}
