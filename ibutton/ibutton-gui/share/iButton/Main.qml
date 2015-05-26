import QtQuick 2.0
import Ubuntu.Components 1.2

/*!
    \brief MainView with a Label and Button elements.
*/

MainView {
    objectName: "mainView"
    applicationName: "ibutton.mh"

    width: units.gu(100)
    height: units.gu(75)

    Page {
        id: page2
        title: i18n.tr("iButton")

        Column {
            spacing: units.gu(1)
            anchors {
                margins: units.gu(2)
                fill: parent
            }

            Button {
                objectName: "button"
                width: parent.width
                text: i18n.tr("Connect")
                onClicked: app.connect()
            }

            Button {
                objectName: "button"
                width: parent.width
                text: i18n.tr("Disconnect")
                onClicked: app.disconnect()
            }

            Label {
                id: labelState
                objectName: "label"
                text: app.state
            }

            Button {
                objectName: "button"
                width: parent.width
                text: i18n.tr("Update Status")
                onClicked: app.update()
            }

            Label {
                id: labelStatus
                objectName: "label"
                text: status
            }

        }
    }

}

