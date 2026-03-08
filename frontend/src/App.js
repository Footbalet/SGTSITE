import { Component } from 'react';
import './styles.scss'
import { Baseplate } from './components/frame/baseplate'

class App extends Component {

    constructor(props) {
        super(props);
        const queryString = window.location.hash.substring(1);
        const hashParams = queryString.split('&')
        const hash = {}
        hashParams.map(el => {
            const data = el.split('=')
            hash[data[0]] = data[1]
            return 0
        })

        let page = hash['page'] ? hash['page'] : undefined
        let sort = hash['sortIndex'] || null;
        let order = 'ASC'
        if (sort !== null && sort.charAt(0) === '-') {
            order = 'DESC'
            sort = sort.slice(1, sort.length)
        }
        let filtersData = {}

        for (const [key, value] of Object.entries(hash)) {
            if (key !== 'sortIndex' && key !== 'page')
                filtersData[key] = value
        }

        if (localStorage.getItem('key_to_my_heart') == null){
            localStorage.setItem('key_to_my_heart', 'no_key')
        }

        var key_to_site = localStorage.getItem('key_to_my_heart') || 'no_key'

        this.state = {
            isStuff: key_to_site !== 'no_key',
            data: null,
            page: page,
            sortIndex: sort,
            sortOrder: order,
            filters: filtersData,
            message: null,
        };
    }

    render() {
        return Baseplate(this)
    }
}

export default App;