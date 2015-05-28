import QtQuick 2.0
import QtQuick.Layouts 1.1
import QtQuick.Controls 1.3
import Ubuntu.Components 1.2
import GoExtensions 1.0


/*!
    \brief MainView with a Label and Button elements.
                                    */
MainView {
    objectName: "main"
    applicationName: "ibutton.mh"

    width: units.gu(100)
    height: units.gu(75)

    App {
        id: app
    }

    Status {
        id: status
        time: "ROFL"
        app: app
    }

    Page {
        title: i18n.tr("iButton")

        ColumnLayout {

            GridLayout {
                columns: 3

                GroupBox {
                    title: "Status"

                    GridLayout {
                        columns: 2
                        rowSpacing: units.gu(1)
                        columnSpacing: units.gu(1)
                        anchors {
                            margins: units.gu(2)
                        }

                        Label {
                            text: "Button time"
                        }
                        Label {
                            text: status.time
                        }

                        Label {
                            text: "Sample rate"
                        }
                        Label {
                            text: status.rate
                        }

                        Label {
                            text: "Resolution"
                        }
                        Label {
                            text: status.resolution
                        }

                        Button {
                            text: i18n.tr("Update Status")
                            onClicked: status.update()
                        }
                        Button {
                            text: i18n.tr("Clear Memory")
                            onClicked: app.clear()
                        }
                    }
                }

                GroupBox {
                    title: "Connection"

                    ColumnLayout {
                        Label {
                            text: app.state
                        }

                        Button {
                            text: i18n.tr("Connect")
                            onClicked: app.connect()
                        }

                        Button {
                            text: i18n.tr("Disconnect")
                            onClicked: app.disconnect()
                        }
                    }
                }

                GroupBox {
                    title: "Mission"

                    GridLayout {
                        columns: 2
                        rowSpacing: units.gu(1)
                        columnSpacing: units.gu(1)
                        anchors {
                            margins: units.gu(2)
                        }

                        Label {
                            text: "Status"
                        }
                        Label {
                            text: status.missionProgress
                        }

                        Label {
                            text: "Started"
                        }
                        Label {
                            text: status.startedTime
                        }

                        Label {
                            text: "Sample count"
                        }
                        Label {
                            text: status.count
                        }

                        Button {
                            text: i18n.tr("Start")
                            onClicked: app.start()
                        }
                        Button {
                            text: i18n.tr("Stop")
                            onClicked: app.stop()
                        }
                }
}
                                }

                GroupBox {
                    title: "Mission Log"


                    Component {
                        id: sampleDelegate
                        Row {
                            spacing: 10
                            Text {
                                text: app.sampleTime(index)
                            }
                            Text {
                                text: app.sampleTemp(index)
                            }
                        }
                    }

                    Column {
                        Button {
                            text: i18n.tr("Read Log")
                            onClicked: app.readLog()
                        }

                        ListView {
                            width: units.gu(90)
                            height: units.gu(30)
                            model: app.samples.len
                            delegate: sampleDelegate
                        }
                    }

            }
        }
    }
}
