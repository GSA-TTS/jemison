//nolint:all
package main

import (
	"log"
	"strings"
	"testing"

	"github.com/GSA-TTS/jemison/internal/util"
	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func TestGetTitle(t *testing.T) {
	html := `
	<html>
		<head>
			<title>Alice was here</title>
		</head>
		<body>
			<title>Bob was here, too</title>
		</body>
	</html>
	`
	reader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		t.Error(err)
	}

	title := "whatevs"
	doc.Find("head title").Each(func(ndx int, sel *goquery.Selection) {
		title = sel.Text()
	})
	assert.Equal(t, "Alice was here", title)
}

var html string = `
<!DOCTYPE html>

<html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta http-equiv="X-UA-Compatible" content="IE=edge">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <meta http-equiv="Content-Security-Policy" 
              content="default-src 'self';
                    connect-src https://www.google-analytics.com https://gov1.siteintercept.qualtrics.com ws://localhost:8080/;
                    frame-src https://feedback.gsa.gov/;
                    img-src 'self' https://gov1.siteintercept.qualtrics.com https://*.amazonaws.com/uploads/form/logo/3110/logo_square_Vector80w(2).jpg;
                    object-src 'none';
                    script-src 'self' 'unsafe-inline' https://gov1.siteintercept.qualtrics.com https://*.gov1.siteintercept.qualtrics.com https://dap.digitalgov.gov https://www.google-analytics.com https://www.googletagmanager.com http://search.usa.gov;
                    style-src 'self' 'unsafe-inline';
                    worker-src 'none';">
        <meta name="Audit submission resources" description="Tools and resources for successfully submiting your single audit package to the Federal Audit Clearinghouse.">
        <link rel="icon" type="image/x-icon" href="/assets/img/favicons/icon_logo.svg" />
        <link rel="stylesheet" href="/assets/css/styles.css" media="all" type="text/css">

        
        

<!--BEGIN QUALTRICS WEBSITE FEEDBACK SNIPPET-->
<script type='text/javascript'>
(function(){var g=function(e,h,f,g){
this.get=function(a){for(var a=a+"=",c=document.cookie.split(";"),b=0,e=c.length;b<e;b++){for(var d=c[b];" "==d.charAt(0);)d=d.substring(1,d.length);if(0==d.indexOf(a))return d.substring(a.length,d.length)}return null};
this.set=function(a,c){var b="",b=new Date;b.setTime(b.getTime()+6048E5);b="; expires="+b.toGMTString();document.cookie=a+"="+c+b+"; path=/; "};
this.check=function(){var a=this.get(f);if(a)a=a.split(":");else if(100!=e)"v"==h&&(e=Math.random()>=e/100?0:100),a=[h,e,0],this.set(f,a.join(":"));else return!0;var c=a[1];if(100==c)return!0;switch(a[0]){case "v":return!1;case "r":return c=a[2]%Math.floor(100/c),a[2]++,this.set(f,a.join(":")),!c}return!0};
this.go=function(){if(this.check()){var a=document.createElement("script");a.type="text/javascript";a.src=g;document.body&&document.body.appendChild(a)}};
this.start=function(){var t=this;"complete"!==document.readyState?window.addEventListener?window.addEventListener("load",function(){t.go()},!1):window.attachEvent&&window.attachEvent("onload",function(){t.go()}):t.go()};};
try{(new g(100,"r","QSI_S_ZN_cN4gyJuBGhPsmcx","https://zncn4gyjubghpsmcx-cemgsa.gov1.siteintercept.qualtrics.com/SIE/?Q_ZID=ZN_cN4gyJuBGhPsmcx")).start()}catch(i){}})();
</script><div id='ZN_cN4gyJuBGhPsmcx'><!--DO NOT REMOVE-CONTENTS PLACED HERE--></div>
<!--END WEBSITE FEEDBACK SNIPPET-->
        
        <title>Audit submission resources</title>
    </head>

    <body>
		<title>lock</title>
		</body>
		</html>
	`

func TestGetTitle2(t *testing.T) {

	reader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		t.Error(err)
	}

	title := ""
	doc.Find("title").Each(func(ndx int, sel *goquery.Selection) {
		if title == "" {
			title = strings.ToLower(sel.Text())
		}
	})
	assert.Equal(t, "audit submission resources", title)

}

func TestGetTitle3(t *testing.T) {

	reader := strings.NewReader(html)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		t.Error(err)
	}

	// The FAC webpages are just malformed for parsing.
	title := ""
	doc.Find("script").Remove()
	doc.Find("head").Each(func(i int, s *goquery.Selection) {
		stripped := util.CollapseWhitespace(s.Find("title").Text())
		if stripped != "" {
			log.Println("setting title to", stripped)
			title = stripped
		}
	})
	assert.Equal(t, "audit submission resources", title)
}
