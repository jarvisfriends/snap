import os

def update(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        c = f.read()
    
    # tabs_test.go
    c = c.replace('if len(tabs.Pages) != 3 {', 'if len(tabs.Pages) != 2 {')
    c = c.replace('expected 3 default pages', 'expected 2 default pages')
    c = c.replace('tabs.ActiveIndex = 1\n', 'tabs.Pages = append(tabs.Pages, Page{ID: "C", Title: "C"})\n\ttabs.ActiveIndex = 1\n')
    
    # topnav_test.go
    c = c.replace('if len(nav.Pages) != 3 {', 'if len(nav.Pages) != 2 {')
    c = c.replace('nav.ActiveIndex = 1\n', 'nav.Pages = append(nav.Pages, Page{ID: "C", Title: "C"})\n\tnav.ActiveIndex = 1\n')
    
    # navigation_test.go
    c = c.replace("after 'down' ActiveIndex = 0; want 1", "after 'down' ActiveIndex = 0; want 1")
    c = c.replace('tabs.ActiveIndex = 2\n', 'tabs.Pages = append(tabs.Pages, Page{ID: "C", Title: "C"})\n\ttabs.ActiveIndex = 2\n')
    
    # render_check_test.go -> fix down from index 0 expecting 1 (which used to be metrics, now is settings)
    # the test expects picked = "metrics" when Enter is pressed after Down.
    # We should just make demoPages() return Home and Metrics instead of Home and Settings?
    # No, demoPages is in main.go of the example. It's actually got 5 pages!
    # Wait, in render_check_test.go, why did down move to 0 instead of 1?
    # Because settingsIdx was 2 but len was 2... wait, the demo uses its own pages!
    # But newDemo uses navStyles() which uses New() which uses the default 2 pages!
    # Then main.go calls SetPages to override them?
    # Wait, let's look at examples/navigation/main.go
    
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(c)

for f in ['navigation/tabs_test.go', 'navigation/topnav_test.go', 'navigation/navigation_test.go']:
    if os.path.exists(f):
        update(f)
