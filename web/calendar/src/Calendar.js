import React, {Component} from 'react';

import moment from 'moment';

import FullCalendar from 'fullcalendar-reactwrapper';

import './Calendar.css';
import 'fullcalendar-reactwrapper/dist/css/fullcalendar.min.css';


class Calendar extends Component {
  constructor(props) {
    super(props);
    this.state = {
      windowHeight: window.innerHeight,
      calendarEvents: [],
    };
  }

  getCalendarEvents(start, end, tz, callback) {
    const dataType = 'alerts';
    fetch(
      'https://newshound.jprbnsn.com/svc/newshound-api/v1/find_' +
        dataType +
        '/' +
        moment(start).format('YYYY-MM-DD') +
        '/' +
        moment(end).format('YYYY-MM-DD'),
    )
      .then(res => res.json())
      .then(
        results => {
          let events = [];
          for (var i = 0; i < results.length; i++) {
            const result = results[i];
            events.push({
              title: result.sender+'\n\n'+result.subject,
              start: result.timestamp,
              obj_id: result.id,
              allDay: false,
              className: result.sender.replace(/(\s|\.|!)/g, '').toLowerCase(),
            });
          }
          callback(events);
        },
        error => {
          console.log(error);
        },
      );
  }

  viewRender(view, element) {
    console.log('V', view);
    console.log('E', element);
  }

  render() {
    return (
      <div id="cal-container">
        <FullCalendar
          theme={false}
          ignoreTimezone={false}
          allDaySlot={false}
          editable={false}
          defaultView="agendaWeek"
          contentHeight={Math.max(this.windowHeight - 400, 500)}
          header={{
            left: 'today prev,next',
            center: '',
            right: '',
          }}
          events={this.getCalendarEvents}
          viewRender={this.viewRender}
        />
      </div>
    );
  }
}

export default Calendar;
