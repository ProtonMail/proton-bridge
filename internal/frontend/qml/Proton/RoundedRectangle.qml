// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

import QtQuick 2.8


Rectangle {
    id: root

    color: "transparent"

    property color fillColor        : Style.currentStyle.background_norm
    property color strokeColor      : Style.currentStyle.background_strong
    property real strokeWidth       : 1

    property real radiusTopLeft     : 10
    property real radiusBottomLeft  : 10
    property real radiusTopRight    : 10
    property real radiusBottomRight : 10

    function paint() {
        canvas.requestPaint()
    }

    onFillColorChanged         : root.paint()
    onStrokeColorChanged       : root.paint()
    onStrokeWidthChanged       : root.paint()
    onRadiusTopLeftChanged     : root.paint()
    onRadiusBottomLeftChanged  : root.paint()
    onRadiusTopRightChanged    : root.paint()
    onRadiusBottomRightChanged : root.paint()


    Canvas {
        id: canvas
        anchors.fill: root

        onPaint: {
            var ctx = getContext("2d")
            ctx.clearRect(0, 0, canvas.width, canvas.height);
            ctx.fillStyle   = root.fillColor
            ctx.strokeStyle = root.strokeColor
            ctx.lineWidth   = root.strokeWidth
            var dimensions = {
                x: ctx.lineWidth,
                y: ctx.lineWidth,
                w: canvas.width-2*ctx.lineWidth,
                h: canvas.height-2*ctx.lineWidth,
            }
            var radius = {
                tl: root.radiusTopLeft,
                tr: root.radiusTopRight,
                bl: root.radiusBottomLeft,
                br: root.radiusBottomRight,
            }

            root.roundRect(
                ctx,
                dimensions,
                radius, true, true
            )
        }
    }

    // adapted from: https://stackoverflow.com/questions/1255512/how-to-draw-a-rounded-rectangle-on-html-canvas/3368118#3368118
    function roundRect(ctx, dim, radius, fill, stroke) {
        if (typeof stroke == 'undefined') {
            stroke = true;
        }
        if (typeof radius === 'undefined') {
            radius = 5;
        }
        if (typeof radius === 'number') {
            radius = {tl: radius, tr: radius, br: radius, bl: radius};
        } else {
            var defaultRadius = {tl: 0, tr: 0, br: 0, bl: 0};
            for (var side in defaultRadius) {
                radius[side] = radius[side] || defaultRadius[side];
            }
        }
        ctx.beginPath();
        ctx.moveTo(dim.x + radius.tl, dim.y);
        ctx.lineTo(dim.x + dim.w - radius.tr, dim.y);
        ctx.quadraticCurveTo(dim.x + dim.w, dim.y, dim.x + dim.w, dim.y + radius.tr);
        ctx.lineTo(dim.x + dim.w, dim.y + dim.h - radius.br);
        ctx.quadraticCurveTo(dim.x + dim.w, dim.y + dim.h, dim.x + dim.w - radius.br, dim.y + dim.h);
        ctx.lineTo(dim.x + radius.bl, dim.y + dim.h);
        ctx.quadraticCurveTo(dim.x, dim.y + dim.h, dim.x, dim.y + dim.h - radius.bl);
        ctx.lineTo(dim.x, dim.y + radius.tl);
        ctx.quadraticCurveTo(dim.x, dim.y, dim.x + radius.tl, dim.y);
        ctx.closePath();
        if (fill) {
            ctx.fill();
        }
        if (stroke) {
            ctx.stroke();
        }
    }

    Component.onCompleted: root.paint()
}

